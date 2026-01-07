package image

import (
	"bytes"
	"context"
	"fmt"

	"image-processing-service/internal/adapters/monitoring"
	"image-processing-service/internal/domain/image"
	"image-processing-service/internal/ports"
)

type SyncTransformInput struct {
	ImageID image.ImageID
	Spec    image.TransformationSpec
}

type TransformOutput struct {
	ID         string `json:"id"`
	VariantKey string `json:"variant_key"`
	MimeType   string `json:"mime_type"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Size       int64  `json:"size"`
}

type TransformImageSyncUseCase struct {
	imageRepo ports.ImageRepository
	storage   ports.ObjectStorage
	processor ports.ImageProcessor
}

func NewTransformImageSyncUseCase(
	imageRepo ports.ImageRepository,
	storage ports.ObjectStorage,
	processor ports.ImageProcessor,
) *TransformImageSyncUseCase {
	return &TransformImageSyncUseCase{
		imageRepo: imageRepo,
		storage:   storage,
		processor: processor,
	}
}

func (uc *TransformImageSyncUseCase) Execute(ctx context.Context, input SyncTransformInput) (*TransformOutput, error) {
	// 1. Generate spec hash for deduplication
	specHash, err := input.Spec.Hash()
	if err != nil {
		return nil, fmt.Errorf("failed to hash transformation spec: %w", err)
	}

	// 2. Check if variant already exists
	existing, err := uc.imageRepo.GetVariantBySpecHash(ctx, input.ImageID, specHash)
	if err == nil && existing != nil {
		monitoring.RecordTransformation("sync", "success")
		return &TransformOutput{
			ID:         existing.ID.String(),
			VariantKey: existing.VariantKey,
			MimeType:   existing.MimeType,
			Width:      existing.Width,
			Height:     existing.Height,
			Size:       existing.Size,
		}, nil
	}

	// 3. Get original image metadata to find storage key
	img, err := uc.imageRepo.GetByID(ctx, input.ImageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image metadata: %w", err)
	}
	if img == nil {
		return nil, fmt.Errorf("image not found: %s", input.ImageID)
	}

	// 4. Download original image
	srcReader, err := uc.storage.Get(ctx, img.OriginalKey)
	if err != nil {
		return nil, fmt.Errorf("failed to download original image: %w", err)
	}
	defer func() {
		_ = srcReader.Close()
	}()

	// 5. Transform image
	processed, err := uc.processor.Transform(ctx, srcReader, &input.Spec)
	if err != nil {
		monitoring.RecordTransformation("sync", "failure")
		return nil, fmt.Errorf("transformation failed: %w", err)
	}

	// 6. Upload variant
	ext := uc.getExtension(processed.MimeType)
	variantKey := fmt.Sprintf("variants/%s/%s%s", input.ImageID, specHash, ext)

	_, err = uc.storage.Put(ctx, variantKey, bytes.NewReader(processed.Data), processed.MimeType, processed.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to upload variant: %w", err)
	}

	return uc.saveAndReturn(ctx, img.ID, variantKey, specHash, processed)
}

func (uc *TransformImageSyncUseCase) saveAndReturn(ctx context.Context, imageID image.ImageID, key, hash string, processed *ports.ProcessedImage) (*TransformOutput, error) {
	// 7. Save variant metadata
	variant, err := image.NewVariant(key, hash, processed.MimeType, processed.Size, processed.Width, processed.Height)
	if err != nil {
		return nil, fmt.Errorf("failed to create variant domain object: %w", err)
	}

	err = uc.imageRepo.SaveVariant(ctx, imageID, variant)
	if err != nil {
		return nil, fmt.Errorf("failed to save variant metadata: %w", err)
	}

	return &TransformOutput{
		ID:         variant.ID.String(),
		VariantKey: variant.VariantKey,
		MimeType:   variant.MimeType,
		Width:      variant.Width,
		Height:     variant.Height,
		Size:       variant.Size,
	}, nil
}

func (uc *TransformImageSyncUseCase) getExtension(mimeType string) string {
	switch mimeType {
	case "image/jpeg", "jpeg", "jpg":
		return ".jpg"
	case "image/png", "png":
		return ".png"
	case "image/webp", "webp":
		return ".webp"
	case "image/gif", "gif":
		return ".gif"
	default:
		return ""
	}
}
