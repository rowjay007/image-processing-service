package image

import (
	"context"
	"fmt"
	"time"

	"image-processing-service/internal/domain/image"
	"image-processing-service/internal/ports"

	"github.com/google/uuid"
)

type AsyncTransformOutput struct {
	ID string
}

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

type AsyncTransformInput struct {
	ImageID image.ImageID
	Spec    image.TransformationSpec
}

func (uc *AsyncTransformImageUseCase) Execute(ctx context.Context, input AsyncTransformInput) (*AsyncTransformOutput, error) {
	// 1. Validate Image exists
	img, err := uc.imageRepo.GetByID(ctx, input.ImageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}
	if img == nil {
		return nil, fmt.Errorf("image not found: %s", input.ImageID)
	}

	// 2. Hash spec
	specHash, err := input.Spec.Hash()
	if err != nil {
		return nil, fmt.Errorf("failed to hash spec: %w", err)
	}

	// 3. Create Job ID
	jobID := uuid.New().String()

	// 4. Publish to Queue
	job := &ports.TransformJob{
		JobID:     jobID,
		ImageID:   string(img.ID),
		OwnerID:   string(img.OwnerID),
		Spec:      &input.Spec,
		SpecHash:  specHash,
		CreatedAt: time.Now().UTC(),
	}

	if err := uc.queue.Publish(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to publish job: %w", err)
	}

	return &AsyncTransformOutput{ID: jobID}, nil
}
