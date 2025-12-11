package processor

import (
	"context"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"

	domainImage "image-processing-service/internal/domain/image"
	"image-processing-service/internal/ports"
)

type StdLibImageProcessor struct{}

func NewStdLibImageProcessor() *StdLibImageProcessor {
	return &StdLibImageProcessor{}
}

func (p *StdLibImageProcessor) ExtractMetadata(ctx context.Context, reader io.Reader) (*ports.ImageMetadata, error) {
	// We need to decode config, not the whole image, to be fast.
	config, format, err := image.DecodeConfig(reader)
	if err != nil {
		return nil, err
	}

	mimeType := "application/octet-stream"
	switch format {
	case "jpeg":
		mimeType = "image/jpeg"
	case "png":
		mimeType = "image/png"
	case "gif":
		mimeType = "image/gif"
	case "webp":
		mimeType = "image/webp" // stdlib might not support webp decode config without x/image?
		// _ "golang.org/x/image/webp" // We might need this if we want webp support
	}

	return &ports.ImageMetadata{
		Width:    config.Width,
		Height:   config.Height,
		MimeType: mimeType,
		Size:     0, // Size is not available from reader decode
	}, nil
}

func (p *StdLibImageProcessor) Transform(ctx context.Context, srcReader io.Reader, spec *domainImage.TransformationSpec) (*ports.ProcessedImage, error) {
	return nil, errors.New("transform not implemented in stdlib processor (use bimg)")
}
