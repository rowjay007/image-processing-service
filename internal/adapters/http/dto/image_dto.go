package dto

import "image"

type ImageMetadataResponse struct {
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

type UploadResponse struct {
	ID          string                `json:"id"`
	OriginalURL string                `json:"original_url"`
	Metadata    ImageMetadataResponse `json:"metadata"`
}

type TransformResponse struct {
	ID         string `json:"id"`
	VariantKey string `json:"variant_key"`
	MimeType   string `json:"mime_type"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Size       int64  `json:"size"`
}

type ListImagesResponse struct {
	Images []*image.Image `json:"images"`
	Total  int            `json:"total"`
}
