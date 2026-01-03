package image

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"image-processing-service/internal/domain/image"
	"image-processing-service/internal/ports"
)

type GetImageUseCase struct {
	repo  ports.ImageRepository
	cache ports.Cache
}

func NewGetImageUseCase(repo ports.ImageRepository, cache ports.Cache) *GetImageUseCase {
	return &GetImageUseCase{
		repo:  repo,
		cache: cache,
	}
}

func (uc *GetImageUseCase) Execute(ctx context.Context, id image.ImageID) (*image.Image, error) {
	// 1. Try Cache
	cacheKey := fmt.Sprintf("image:%s", id)
	cachedVal, err := uc.cache.Get(ctx, cacheKey)
	if err == nil && cachedVal != "" {
		var cachedImage image.Image
		if jsonErr := json.Unmarshal([]byte(cachedVal), &cachedImage); jsonErr == nil {
			return &cachedImage, nil
		}
	}

	// 2. Fetch from Repo
	img, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}
	
	// 3. Set Cache (Async or Sync? Sync for now, but don't block too long)
	if img != nil {
		if bytes, err := json.Marshal(img); err == nil {
			// TTL: 1 hour? Configurable?
			_ = uc.cache.Set(ctx, cacheKey, string(bytes), 1*time.Hour)
		}
	}

	return img, nil
}
