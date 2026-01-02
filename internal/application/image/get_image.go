package image

import (
	"context"
	"fmt"

	"image-processing-service/internal/domain/image"
	"image-processing-service/internal/ports"
)

type GetImageUseCase struct {
	repo ports.ImageRepository
}

func NewGetImageUseCase(repo ports.ImageRepository) *GetImageUseCase {
	return &GetImageUseCase{
		repo: repo,
	}
}

func (uc *GetImageUseCase) Execute(ctx context.Context, id image.ImageID) (*image.Image, error) {
	img, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}
	
	// Note: PostgresImageRepository GetByID currently fetches variants too in a second query? 
	// Let's check repository implementation.
	// Yes, I implemented it to fetch variants.
	
	return img, nil
}
