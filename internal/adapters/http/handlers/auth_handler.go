package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"image-processing-service/internal/adapters/auth"
	"image-processing-service/internal/adapters/http/dto"
	appAuth "image-processing-service/internal/application/auth"
)

type AuthHandler struct {
	registerUC *appAuth.RegisterUserUseCase
	loginUC    *appAuth.LoginUserUseCase
	hasher     *auth.BcryptPasswordHasher
}

func NewAuthHandler(registerUC *appAuth.RegisterUserUseCase, loginUC *appAuth.LoginUserUseCase, hasher *auth.BcryptPasswordHasher) *AuthHandler {
	return &AuthHandler{
		registerUC: registerUC,
		loginUC:    loginUC,
		hasher:     hasher,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := h.hasher.Hash(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process password"})
		return
	}

	user, err := h.registerUC.Execute(c.Request.Context(), req.Username, hashedPassword)
	if err != nil {
		if err == appAuth.ErrUserAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": dto.UserResponse{
			ID:       string(user.ID),
			Username: user.Username,
		},
		"message": "User registered successfully. Please login.",
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.loginUC.Execute(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if err == appAuth.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		User: dto.UserResponse{
			ID:       string(user.ID),
			Username: user.Username,
		},
		Token: token,
	})
}
