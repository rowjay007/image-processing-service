package dto

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
