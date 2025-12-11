package image

import (
	"errors"
	"time"

	"image-processing-service/internal/domain/user"

	"github.com/google/uuid"
)

type ImageID string

type Image struct {
	ID          ImageID
	OwnerID     user.UserID
	Filename    string
	OriginalKey string 
	Size        int64
	MimeType    string
	Width       int
	Height      int
	Variants    []Variant
	CreatedAt   time.Time
}

var (
	ErrInvalidImageID  = errors.New("invalid image ID")
	ErrInvalidOwnerID  = errors.New("invalid owner ID")
	ErrInvalidFilename = errors.New("invalid filename")
)

func New(ownerID user.UserID, filename, originalKey, mimeType string, size int64, width, height int) (*Image, error) {
	if ownerID == "" {
		return nil, ErrInvalidOwnerID
	}
	if filename == "" {
		return nil, ErrInvalidFilename
	}
	if originalKey == "" {
		return nil, errors.New("original key cannot be empty")
	}

	return &Image{
		ID:          ImageID(uuid.New().String()),
		OwnerID:     ownerID,
		Filename:    filename,
		OriginalKey: originalKey,
		Size:        size,
		MimeType:    mimeType,
		Width:       width,
		Height:      height,
		Variants:    make([]Variant, 0),
		CreatedAt:   time.Now().UTC(),
	}, nil
}

func (i *Image) AddVariant(v Variant) bool {
	for _, existing := range i.Variants {
		if existing.SpecHash == v.SpecHash {
			return false
		}
	}
	i.Variants = append(i.Variants, v)
	return true
}

func (i *Image) GetVariant(specHash string) *Variant {
	for _, v := range i.Variants {
		if v.SpecHash == specHash {
			return &v
		}
	}
	return nil
}
