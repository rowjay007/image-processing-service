package image

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Variant represents a transformed version of an original image.
type Variant struct {
	ID         uuid.UUID
	VariantKey string // Storage key (path)
	SpecHash   string // Hash of the transformation spec
	Size       int64  // File size in bytes
	MimeType   string
	Width      int
	Height     int
	CreatedAt  time.Time
}

var (
	ErrInvalidVariantKey = errors.New("invalid variant key")
	ErrInvalidSpecHash   = errors.New("invalid spec hash")
)

// NewVariant creates a new Variant instance.
func NewVariant(key, specHash, mimeType string, size int64, width, height int) (*Variant, error) {
	if key == "" {
		return nil, ErrInvalidVariantKey
	}
	if specHash == "" {
		return nil, ErrInvalidSpecHash
	}

	return &Variant{
		ID:         uuid.New(),
		VariantKey: key,
		SpecHash:   specHash,
		Size:       size,
		MimeType:   mimeType,
		Width:      width,
		Height:     height,
		CreatedAt:  time.Now().UTC(),
	}, nil
}
