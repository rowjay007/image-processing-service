package image

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"image-processing-service/internal/domain/image"
	"image-processing-service/internal/ports"
)

type AsyncTransformImageUseCase struct {
	imageRepo ports.ImageRepository
	queue     ports.Queue
}

func NewAsyncTransformImageUseCase(imageRepo ports.ImageRepository, queue ports.Queue) *AsyncTransformImageUseCase {
	return &AsyncTransformImageUseCase{
		imageRepo: imageRepo,
		queue:     queue,
	}
}

type TransformInput struct {
	ImageID image.ImageID
	Spec    image.TransformationSpec
}

func (uc *AsyncTransformImageUseCase) Execute(ctx context.Context, input TransformInput) (*image.Variant, error) {
	// 1. Validate Spec
	// The spec structs tags handle some, but we might want logic check.
	// Spec validation logic is ideally in domain.
	
	specHash, err := input.Spec.Hash()
	if err != nil {
		return nil, fmt.Errorf("invalid spec: %w", err)
	}

	// 2. Check if Image Exists
	img, err := uc.imageRepo.GetByID(ctx, input.ImageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}
	if img == nil {
		return nil, fmt.Errorf("image not found")
	}

	// 3. Check for Existing Variant (Idempotency)
	existing, err := uc.imageRepo.GetVariantBySpecHash(ctx, input.ImageID, specHash)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing variant: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	// 4. Create Pending Variant
	// We create a "pending" variant entry or just fire the job?
	// If we don't create an entry, the user can't poll for it easily unless we return a JobID.
	// But our domain `Variant` doesn't explicitly have a "Status" field in the entity I defined earlier?
	// Let's check `variant.go`. I recall implemented it.
	// If no status, we might just assume if it exists in DB it's done?
	// Or we might risk race conditions.
	// For MVP, we can return a "JobID" (which might be the variant ID) and the user polls `GET /images/{id}`.
	// If the variant isn't there, it's not done.
	// But we need to know if it's *processing* or *failed*.
	// I'll assume for now we just fire the job. The worker will create the variant in DB upon completion.
	// Wait, `SyncTransform` would create it immediately. `Async` should probably just queue.
	// BUT, if we want to return the Variant ID so the user can look for it later?
	// We can pre-calculate the Variant ID (deterministically) or generate a new one.
	// `NewVariant` generates a random UUID.
	// If I don't save it now, I can't return it easily to look up later unless I save a "Job" entity.
	// The prompt didn't ask for a Job entity.
	// Minimal approach:
	// Return a JobID (which is just a UUID to track this request).
	// OR create the Variant with "Status: Pending" if domain allows.
	// `Variant` struct in `variant.go`:
	/*
	type Variant struct {
		ID uuid.UUID
		...
		VariantKey string
		...
	}
	*/
	// Use case returns `*image.Variant` or error.
	// If I return nil, it implies failure.
	// I should probably change signature to return `jobID` or `*Variant` (pending).
	// Let's create a "Pending" variant in DB if we want to track it, but we need fields for that (Status).
	// If I didn't add Status to Variant, I can't distinct pending vs done.
	// Check `migrations/003_create_variants_table.sql`. (I haven't created it yet! I created 002!)
	// Ah, I need to create 003 migration.
	// And I should add a status column to Variant.
	
	// Since I can't easily change the domain without "Code Cleanup" style effort,
	// I will just queue the job and return "Accepted".
	// But the function signature returns `*image.Variant`.
	// I'll allow returning `nil` with no error to indicate "Accepted, check back later"?
	// Or better: Return a Variant struct that has "Status" but maybe it's not persisted yet?
	// Actually, the `TransformJob` needs the ID of the variant to save it as?
	// Worker needs to know what ID to save.
	// So I should Generate ID -> Queue Job with ID -> Return ID.
	// The Variant won't be in DB yet.
	// When user gets image, they won't see it until worker finishes. This is fine for MVP.
	
	variantID := uuid.New()
	
	job := ports.TransformJob{
		JobID:     uuid.New().String(),
		ImageID:   string(input.ImageID),
		VariantID: variantID.String(),
		Spec:      &input.Spec,
		OwnerID:   string(img.OwnerID),
		CreatedAt: time.Now(),
	}

	if err := uc.queue.Publish(ctx, &job); err != nil {
		return nil, fmt.Errorf("failed to queue job: %w", err)
	}

	// Make a temporary variant object to return, so user knows the ID?
	// Or just return nil and let controller say "Accepted".
	// The Controller likely wants to return 202 Accepted with a Location header or similar.
	// Returns a "Virtual" variant.
	v := &image.Variant{
		ID: variantID,
		// Other fields empty
	}
	return v, nil
}
