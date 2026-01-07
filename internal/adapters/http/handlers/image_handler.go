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

// Upload handles image upload
// @Summary Upload an image
// @Description Upload a new image for processing. Requires multipart/form-data.
// @Tags images
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Image file to upload"
// @Success 201 {object} dto.UploadResponse "Image uploaded successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /images [post]
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
	defer func() {
		_ = file.Close()
	}()

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

// Transform handles image transformation
// @Summary Transform an image
// @Description Apply transformations to an image. Use sync=true for immediate response.
// @Tags images
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Image ID"
// @Param sync query boolean false "Perform transformation synchronously"
// @Param spec body image.TransformationSpec true "Transformation metadata"
// @Success 200 {object} dto.TransformResponse "Transformation result (sync)"
// @Success 202 {object} map[string]interface{} "Transformation accepted (async)"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /images/{id}/transform [post]
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
		"message":    "Transformation accepted",
		"variant_id": result.ID,
	})
}

// Get handles fetching image details
// @Summary Get image details
// @Description Fetch metadata and variants for a specific image
// @Tags images
// @Produce json
// @Security BearerAuth
// @Param id path string true "Image ID"
// @Success 200 {object} image.Image "Image details"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Image not found"
// @Router /images/{id} [get]
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

// List handles listing user images
// @Summary List user images
// @Description List all images owned by the user with pagination
// @Tags images
// @Produce json
// @Security BearerAuth
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for pagination" default(10)
// @Success 200 {object} dto.ListImagesResponse "List of images"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /images [get]
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
