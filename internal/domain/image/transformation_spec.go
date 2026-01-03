package image

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// TransformationSpec defines the desired operations to apply to an image.
// It serves as a value object that determines the unique identity of a variant.
type TransformationSpec struct {
	Resize    *ResizeSpec    `json:"resize,omitempty"`
	Crop      *CropSpec      `json:"crop,omitempty"`
	Rotate    *int           `json:"rotate,omitempty" binding:"omitempty,oneof=0 90 180 270"`
	Flip      bool           `json:"flip,omitempty"`
	Mirror    bool           `json:"mirror,omitempty"`
	Watermark *WatermarkSpec `json:"watermark,omitempty"`
	Quality   *int           `json:"quality,omitempty" binding:"omitempty,min=1,max=100"`
	Format    *string        `json:"format,omitempty" binding:"omitempty,oneof=jpeg png webp gif"`
	Filters   *FilterSpec    `json:"filters,omitempty"`
}

type ResizeSpec struct {
	Width  int `json:"width" binding:"required_with=Height,min=1,max=8000"`
	Height int `json:"height" binding:"required_with=Width,min=1,max=8000"`
}

type CropSpec struct {
	Width  int `json:"width" binding:"required,min=1,max=8000"`
	Height int `json:"height" binding:"required,min=1,max=8000"`
	X      int `json:"x" binding:"min=0"`
	Y      int `json:"y" binding:"min=0"`
}

type WatermarkSpec struct {
	Text      string  `json:"text,omitempty"`
	ImageID   string  `json:"image_id,omitempty"`
	Opacity   float64 `json:"opacity"` // 0.0-1.0
	Gravity   string  `json:"gravity"` // center, north, southeast, etc.
}

type FilterSpec struct {
	Grayscale bool `json:"grayscale,omitempty"`
	Sepia     bool `json:"sepia,omitempty"`
	Blur      int  `json:"blur,omitempty"` // sigma
}

func (s *TransformationSpec) Hash() (string, error) {
	bytes, err := json.Marshal(s)
	if err != nil {
		return "", fmt.Errorf("failed to marshal spec for hashing: %w", err)
	}

	hasher := sha256.New()
	hasher.Write(bytes)
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// String returns the string representation (JSON) of the spec.
func (s *TransformationSpec) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}
