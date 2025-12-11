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
	Rotate    *int           `json:"rotate,omitempty"` // degrees: 90, 180, 270
	Flip      bool           `json:"flip,omitempty"`   // vertical flip
	Mirror    bool           `json:"mirror,omitempty"` // horizontal flip
	Watermark *WatermarkSpec `json:"watermark,omitempty"`
	Quality   *int           `json:"quality,omitempty"` // 1-100
	Format    *string        `json:"format,omitempty"`  // jpeg, png, webp, etc.
	Filters   *FilterSpec    `json:"filters,omitempty"`
}

type ResizeSpec struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type CropSpec struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	X      int `json:"x"`
	Y      int `json:"y"`
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
