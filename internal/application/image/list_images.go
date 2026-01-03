package image

import (
	"context"
	"fmt"

	"image-processing-service/internal/domain/image"
	"image-processing-service/internal/domain/user"
	"image-processing-service/internal/ports"
)

type ListImagesUseCase struct {
	repo  ports.ImageRepository
	cache ports.Cache
}

func NewListImagesUseCase(repo ports.ImageRepository, cache ports.Cache) *ListImagesUseCase {
	return &ListImagesUseCase{
		repo:  repo,
		cache: cache,
	}
}

type ListImagesInput struct {
	OwnerID user.UserID
	Offset  int
	Limit   int
}

type ListImagesOutput struct {
	Images []*image.Image
	Total  int
}

func (uc *ListImagesUseCase) Execute(ctx context.Context, input ListImagesInput) (*ListImagesOutput, error) {
	if input.Limit <= 0 {
		input.Limit = 10
	}
	if input.Limit > 100 {
		input.Limit = 100
	}

	images, total, err := uc.repo.List(ctx, input.OwnerID, input.Offset, input.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	return &ListImagesOutput{
		Images: images,
		Total:  total,
	}, nil
}
