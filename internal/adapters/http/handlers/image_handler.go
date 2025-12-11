package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"image-processing-service/internal/adapters/http/dto"
	appImage "image-processing-service/internal/application/image"
	"image-processing-service/internal/domain/user"
)

type ImageHandler struct {
	uploadUC *appImage.UploadImageUseCase
}

func NewImageHandler(uploadUC *appImage.UploadImageUseCase) *ImageHandler {
	return &ImageHandler{
		uploadUC: uploadUC,
	}
}

func (h *ImageHandler) Upload(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := user.UserID(userIDStr.(string))

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer file.Close()

	input := appImage.UploadInput{
		OwnerID:  userID,
		Filename: filepath.Base(fileHeader.Filename),
		File:     file,
		Size:     fileHeader.Size,
		MimeType: fileHeader.Header.Get("Content-Type"),
	}

	img, err := h.uploadUC.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("upload failed: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, dto.UploadResponse{
		ID:          string(img.ID),
		OriginalURL: img.OriginalKey,
		Metadata: dto.ImageMetadataResponse{
			Size:     img.Size,
			MimeType: img.MimeType,
			Width:    img.Width,
			Height:   img.Height,
		},
	})
}
