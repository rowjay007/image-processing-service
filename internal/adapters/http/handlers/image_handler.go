package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"image-processing-service/internal/adapters/http/dto"
	appImage "image-processing-service/internal/application/image"
	"image-processing-service/internal/domain/image"
	"image-processing-service/internal/domain/user"
)

type ImageHandler struct {
	uploadUC         *appImage.UploadImageUseCase
	asyncTransformUC *appImage.AsyncTransformImageUseCase
	syncTransformUC  *appImage.TransformImageSyncUseCase
	getUC            *appImage.GetImageUseCase
	listUC           *appImage.ListImagesUseCase
}

func NewImageHandler(
	uploadUC *appImage.UploadImageUseCase, 
	asyncTransformUC *appImage.AsyncTransformImageUseCase,
	syncTransformUC *appImage.TransformImageSyncUseCase,
	getUC *appImage.GetImageUseCase,
	listUC *appImage.ListImagesUseCase,
) *ImageHandler {
	return &ImageHandler{
		uploadUC:         uploadUC,
		asyncTransformUC: asyncTransformUC,
		syncTransformUC:  syncTransformUC,
		getUC:            getUC,
		listUC:           listUC,
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

func (h *ImageHandler) Transform(c *gin.Context) {
	// Parse ID
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image id required"})
		return
	}
	
	// Parse Body (Spec)
	var spec image.TransformationSpec
	if err := c.ShouldBindJSON(&spec); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transformation spec"})
		return
	}
	
	imageID := image.ImageID(idStr)
	isSync := c.Query("sync") == "true"

	if isSync {
		input := appImage.SyncTransformInput{
			ImageID: imageID,
			Spec:    spec,
		}
		result, err := h.syncTransformUC.Execute(c.Request.Context(), input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("sync transform failed: %v", err)})
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}

	input := appImage.AsyncTransformInput{
		ImageID: imageID,
		Spec:    spec,
	}
	result, err := h.asyncTransformUC.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("transform failed: %v", err)})
		return
	}
	
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Transformation accepted",
		"variant_id": result.ID,
	})
}

func (h *ImageHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image id required"})
		return
	}

	img, err := h.getUC.Execute(c.Request.Context(), image.ImageID(idStr))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
		return
	}
	if img == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
		return
	}

	c.JSON(http.StatusOK, img)
}

func (h *ImageHandler) List(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := user.UserID(userIDStr.(string))
	
	input := appImage.ListImagesInput{
		OwnerID: userID,
		Offset:  0,
		Limit:   10,
	}
	// TODO: Parse offset/limit from query params if needed
	
	result, err := h.listUC.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list images"})
		return
	}
	
	c.JSON(http.StatusOK, result)
}
