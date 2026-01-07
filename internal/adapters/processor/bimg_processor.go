package processor

import (
	"context"
	"fmt"
	"io"

	"github.com/h2non/bimg"

	"image-processing-service/internal/domain/image"
	"image-processing-service/internal/ports"
)

type BimgProcessor struct{}

func NewBimgProcessor() *BimgProcessor {
	return &BimgProcessor{}
}

func (p *BimgProcessor) Transform(ctx context.Context, srcReader io.Reader, spec *image.TransformationSpec) (*ports.ProcessedImage, error) {
	buffer, err := io.ReadAll(srcReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read source image: %w", err)
	}

	img := bimg.NewImage(buffer)
	options := bimg.Options{}

	// Resize
	if spec.Resize != nil {
		options.Width = spec.Resize.Width
		options.Height = spec.Resize.Height
		options.Embed = true // Keep aspect ratio by default or as needed
	}

	// Crop
	if spec.Crop != nil {
		options.Top = spec.Crop.Y
		options.Left = spec.Crop.X
		options.AreaWidth = spec.Crop.Width
		options.AreaHeight = spec.Crop.Height
		options.Crop = true
	}

	// Rotate
	if spec.Rotate != nil {
		options.Rotate = bimg.Angle(*spec.Rotate)
	}

	// Flip/Mirror
	if spec.Flip {
		options.Flip = true
	}
	if spec.Mirror {
		options.Flop = true
	}

	// Quality
	if spec.Quality != nil {
		options.Quality = *spec.Quality
	}

	// Format
	if spec.Format != nil {
		options.Type = p.toBimgType(*spec.Format)
	}

	// Filters
	if spec.Filters != nil {
		if spec.Filters.Grayscale {
			options.Interpretation = bimg.InterpretationBW
		}
		if spec.Filters.Blur > 0 {
			options.GaussianBlur = bimg.GaussianBlur{
				Sigma: float64(spec.Filters.Blur),
			}
		}
	}

	// Watermark - Not implemented in this version

	newBuffer, err := img.Process(options)
	if err != nil {
		return nil, fmt.Errorf("bimg processing failed: %w", err)
	}

	metadata, err := bimg.Metadata(newBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to get processed image metadata: %w", err)
	}

	return &ports.ProcessedImage{
		Data:     newBuffer,
		MimeType: p.getMimeType(bimg.DetermineImageType(newBuffer)),
		Width:    metadata.Size.Width,
		Height:   metadata.Size.Height,
		Size:     int64(len(newBuffer)),
	}, nil
}

func (p *BimgProcessor) ExtractMetadata(ctx context.Context, reader io.Reader) (*ports.ImageMetadata, error) {
	buffer, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read image for metadata: %w", err)
	}

	size, err := bimg.Size(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to get image size: %w", err)
	}

	return &ports.ImageMetadata{
		Width:    size.Width,
		Height:   size.Height,
		MimeType: p.getMimeType(bimg.DetermineImageType(buffer)),
		Size:     int64(len(buffer)),
	}, nil
}

func (p *BimgProcessor) getMimeType(t bimg.ImageType) string {
	switch t {
	case bimg.JPEG:
		return "image/jpeg"
	case bimg.PNG:
		return "image/png"
	case bimg.WEBP:
		return "image/webp"
	case bimg.GIF:
		return "image/gif"
	case bimg.TIFF:
		return "image/tiff"
	case bimg.PDF:
		return "application/pdf"
	case bimg.SVG:
		return "image/svg+xml"
	case bimg.HEIF:
		return "image/heif"
	case bimg.AVIF:
		return "image/avif"
	default:
		return "application/octet-stream"
	}
}

func (p *BimgProcessor) toBimgType(format string) bimg.ImageType {
	switch format {
	case "jpeg", "jpg":
		return bimg.JPEG
	case "png":
		return bimg.PNG
	case "webp":
		return bimg.WEBP
	case "gif":
		return bimg.GIF
	case "tiff":
		return bimg.TIFF
	case "pdf":
		return bimg.PDF
	case "svg":
		return bimg.SVG
	case "heif":
		return bimg.HEIF
	case "avif":
		return bimg.AVIF
	case "magick":
		return bimg.MAGICK
	default:
		return bimg.UNKNOWN
	}
}
