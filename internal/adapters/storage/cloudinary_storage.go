package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"

	"image-processing-service/internal/config"
)

type CloudinaryStorage struct {
	client *cloudinary.Cloudinary
	config config.CloudinaryConfig
}

func NewCloudinaryStorage(cfg config.CloudinaryConfig) (*CloudinaryStorage, error) {
	cld, err := cloudinary.NewFromParams(cfg.CloudName, cfg.APIKey, cfg.APISecret)
	if err != nil {
		return nil, fmt.Errorf("failed to init cloudinary: %w", err)
	}
	return &CloudinaryStorage{
		client: cld,
		config: cfg,
	}, nil
}

func (s *CloudinaryStorage) Put(ctx context.Context, key string, reader io.Reader, contentType string, size int64) (string, error) {
	params := uploader.UploadParams{
		PublicID:     key,
		Folder:       s.config.Folder,
		Overwrite:    api.Bool(true),
		ResourceType: "image",
	}

	result, err := s.client.Upload.Upload(ctx, reader, params)
	if err != nil {
		return "", fmt.Errorf("cloudinary upload failed: %w", err)
	}

	return result.SecureURL, nil
}

func (s *CloudinaryStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	_, err := s.SignedURL(ctx, key, 1*time.Hour)
	if err != nil {
		return nil, err
	}
	
	return nil, fmt.Errorf("direct get not implemented for cloudinary, use SignedURL")
}

func (s *CloudinaryStorage) SignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	asset, err := s.client.Image(key)
	if err != nil {
		return "", err
	}
	
	url, err := asset.String()
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *CloudinaryStorage) Delete(ctx context.Context, key string) error {
	_, err := s.client.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: key,
	})
	return err
}
