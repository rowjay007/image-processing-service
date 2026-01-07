package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if err := os.MkdirAll(filepath.Clean(basePath), 0750); err != nil {
		return nil, fmt.Errorf("failed to create base path: %w", err)
	}
	return &LocalStorage{basePath: filepath.Clean(basePath)}, nil
}

func (s *LocalStorage) Put(ctx context.Context, key string, reader io.Reader, contentType string, size int64) (string, error) {
	fullPath := filepath.Join(s.basePath, filepath.Clean(key))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0750); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(filepath.Clean(fullPath))
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if _, err := io.Copy(file, reader); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fullPath, nil
}

func (s *LocalStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, filepath.Clean(key))
	file, err := os.Open(filepath.Clean(fullPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

func (s *LocalStorage) SignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	// For local storage, a "signed URL" is just the local path or a mocked URL.
	// In a real system, this would be served by the API.
	return filepath.Join(s.basePath, key), nil
}

func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(s.basePath, key)
	return os.Remove(fullPath)
}
