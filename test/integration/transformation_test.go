package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"image-processing-service/internal/container"
	"image-processing-service/internal/database"
)

func TestSyncTransformationIntegration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test; RUN_INTEGRATION_TESTS not set to true")
	}

	c, err := container.NewContainer()
	require.NoError(t, err)
	defer c.Close()

	migrationDir := "../../migrations"
	err = database.RunMigrations(context.Background(), c.DB, migrationDir)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Setup routes
	authMiddleware := c.AuthMiddleware.Handle()
	r.POST("/images", authMiddleware, c.ImageHandler.Upload)
	r.POST("/images/:id/transform", authMiddleware, c.ImageHandler.Transform)

	// 1. Login to get token
	token := getTestToken(t, r, c)

	// 2. Upload an image
	imageID := uploadTestImage(t, r, token)

	// 3. Request Sync Transformation
	spec := map[string]interface{}{
		"resize": map[string]interface{}{
			"width":  200,
			"height": 200,
		},
		"format":  "webp",
		"quality": 80,
	}
	specJSON, _ := json.Marshal(spec)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/images/%s/transform?sync=true", imageID), bytes.NewBuffer(specJSON))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.NotEmpty(t, result["id"])
	assert.Equal(t, "image/webp", result["mime_type"])
	assert.NotNil(t, result["variant_key"])

	variantID1 := result["id"].(string)

	// 4. Request the SAME Transformation Again (Deduplication Check)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req)

	assert.Equal(t, http.StatusOK, w2.Code)

	var result2 map[string]interface{}
	_ = json.Unmarshal(w2.Body.Bytes(), &result2)

	assert.Equal(t, variantID1, result2["id"].(string), "Expected same variant ID (deduplication)")
}

func getTestToken(t *testing.T, r *gin.Engine, c *container.Container) string {
	username := fmt.Sprintf("testuser_%d", os.Getpid())
	password := "Password123!"

	regBody := fmt.Sprintf(`{"username":"%s", "password":"%s"}`, username, password)
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")

	r.POST("/auth/register", c.AuthHandler.Register)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated && w.Code != http.StatusConflict {
		t.Fatalf("Expected 201 Created or 409 Conflict for register, got %d: %s", w.Code, w.Body.String())
	}

	loginBody := fmt.Sprintf(`{"username":"%s", "password":"%s"}`, username, password)
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")

	r.POST("/auth/login", c.AuthHandler.Login)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for login, got %d: %s", w.Code, w.Body.String())
	}

	var loginResp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &loginResp)
	return loginResp.Token
}

func uploadTestImage(t *testing.T, r *gin.Engine, token string) string {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.png")

	// Create a small 1x1 pixel PNG for testing
	// Red dot PNG
	_, _ = part.Write([]byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90wS\xde\x00\x00\x00\x0cIDAT\x08\xd7\x63\xf8\xff\xff\x3f\x00\x05\xfe\x02\xfe\xdc\x44\x74\x73\x00\x00\x00\x00IEND\xaeB`\x82"))
	_ = writer.Close()

	req, _ := http.NewRequest("POST", "/images", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var resp struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.ID
}
