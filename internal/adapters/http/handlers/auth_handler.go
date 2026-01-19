package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"image-processing-service/internal/adapters/http/dto"
	"image-processing-service/internal/ports"
)

type AuthHandler struct {
	authClient ports.RemoteAuthService
}

func NewAuthHandler(authClient ports.RemoteAuthService) *AuthHandler {
	return &AuthHandler{
		authClient: authClient,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account with username and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration details"
// @Success 201 {object} map[string]interface{} "User registered successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 409 {object} map[string]interface{} "User already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, username, err := h.authClient.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		// We could be more specific with error mapping here
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": dto.UserResponse{
			ID:       userID,
			Username: username,
		},
		"message": "User registered successfully. Please login.",
	})
}

// Login handles user authentication
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.AuthResponse "Login successful"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, username, token, err := h.authClient.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials or auth service error"})
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		User: dto.UserResponse{
			ID:       userID,
			Username: username,
		},
		Token: token,
	})
}

// Me returns the current user profile
// @Summary Get current user profile
// @Description Get details of the currently authenticated user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User profile details"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("userID")
	username, _ := c.Get("username")
	c.JSON(http.StatusOK, gin.H{
		"userId":   userID,
		"username": username,
	})
}
