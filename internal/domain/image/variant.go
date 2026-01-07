package image

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Variant represents a transformed version of an original image.
type Variant struct {
	ID         uuid.UUID `json:"id"`
	VariantKey string    `json:"variant_key"`
	SpecHash   string    `json:"spec_hash"`
	Size       int64     `json:"size"`
	MimeType   string    `json:"mime_type"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	CreatedAt  time.Time `json:"created_at"`
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
