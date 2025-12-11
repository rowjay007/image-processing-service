package image

import (
	"context"
	"fmt"
	"mime/multipart"

	"image-processing-service/internal/domain/image"
	"image-processing-service/internal/domain/user"
	"image-processing-service/internal/ports"
)

type UploadImageUseCase struct {
	imageRepo ports.ImageRepository
	storage   ports.ObjectStorage
	processor ports.ImageProcessor 
}

func NewUploadImageUseCase(imageRepo ports.ImageRepository, storage ports.ObjectStorage, processor ports.ImageProcessor) *UploadImageUseCase {
	return &UploadImageUseCase{
		imageRepo: imageRepo,
		storage:   storage,
		processor: processor,
	}
}

type UploadInput struct {
	OwnerID  user.UserID
	Filename string
	File     multipart.File
	Size     int64
	MimeType string
}

func (uc *UploadImageUseCase) Execute(ctx context.Context, input UploadInput) (*image.Image, error) {
    width, height := 0, 0
    if uc.processor != nil {
        meta, err := uc.processor.ExtractMetadata(ctx, input.File)
        if err != nil {
            return nil, fmt.Errorf("invalid image: %w", err)
        }
        width = meta.Width
        height = meta.Height
        input.MimeType = meta.MimeType
        if _, err := input.File.Seek(0, 0); err != nil {
             return nil, fmt.Errorf("failed to reset file pointer: %w", err)
        }
    }

	tempImg, err := image.New(input.OwnerID, input.Filename, "temp", input.MimeType, input.Size, width, height)
	if err != nil {
		return nil, err
	}
	
	key := fmt.Sprintf("users/%s/images/%s/original", input.OwnerID, tempImg.ID)
	tempImg.OriginalKey = key

	if _, err := uc.storage.Put(ctx, key, input.File, input.MimeType, input.Size); err != nil {
		return nil, fmt.Errorf("storage upload failed: %w", err)
	}

	if err := uc.imageRepo.Save(ctx, tempImg); err != nil {
		_ = uc.storage.Delete(ctx, key)
		return nil, err
	}

	return tempImg, nil
}
