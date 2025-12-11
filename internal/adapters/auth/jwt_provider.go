package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"image-processing-service/internal/config"
	"image-processing-service/internal/domain/user"
	"image-processing-service/internal/ports"
)

type JWTProvider struct {
	config config.JWTConfig
}

func NewJWTProvider(cfg config.JWTConfig) *JWTProvider {
	return &JWTProvider{
		config: cfg,
	}
}

type jwtClaims struct {
	UserID   string `json:"sub"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (p *JWTProvider) GenerateToken(userID user.UserID, username string) (string, error) {
	claims := jwtClaims{
		UserID:   string(userID),
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(p.config.Expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    p.config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(p.config.Secret))
}

func (p *JWTProvider) ValidateToken(tokenString string) (*ports.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(p.config.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
		return &ports.Claims{
			UserID:   claims.UserID,
			Username: claims.Username,
		}, nil
	}

	return nil, errors.New("invalid token")
}
