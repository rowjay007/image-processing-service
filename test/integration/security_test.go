package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityAndRateLimiting(t *testing.T) {
	if os.Getenv("RUN_LIVE_TESTS") != "true" {
		t.Skip("Skipping live security test; RUN_LIVE_TESTS not set to true")
	}
	baseURL := "http://localhost:8080/api/v1"
	client := &http.Client{}

	t.Run("Security Headers", func(t *testing.T) {
		resp, err := client.Get("http://localhost:8080/health")
		if err != nil {
			t.Fatalf("Failed to call health check: %v", err)
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
		assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
		assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "default-src 'none'")
	})

	t.Run("Input Validation Register", func(t *testing.T) {
		registerURL := baseURL + "/auth/register"
		badRegister := map[string]string{
			"username": "ok",
			"password": "123", // too short
		}
		jsonData, _ := json.Marshal(badRegister)
		resp, err := client.Post(registerURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should return 400 for short password")
	})

	t.Run("Rate Limiting Auth", func(t *testing.T) {
		loginURL := baseURL + "/auth/login"

		// The limit is 5 per 15 mins. Let's hit it 6 times.
		for i := 0; i < 6; i++ {
			badLogin := map[string]string{
				"username": fmt.Sprintf("user_%d", i),
				"password": "password123",
			}
			jsonData, _ := json.Marshal(badLogin)
			resp, err := client.Post(loginURL, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatalf("Failed at attempt %d: %v", i, err)
			}
			defer func() {
			_ = resp.Body.Close()
		}()

			if i == 5 {
				assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode, "Should be rate limited on 6th attempt")
			} else {
				// Expecting 401 because user doesn't exist, but NOT 429
				assert.NotEqual(t, http.StatusTooManyRequests, resp.StatusCode, fmt.Sprintf("Should NOT be rate limited on attempt %d", i))
			}
		}
	})
}
