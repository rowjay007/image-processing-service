package auth

import (
	"context"
	"errors"

	"image-processing-service/internal/domain/user"
	"image-processing-service/internal/ports"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type PasswordHasher interface {
	Compare(hash, password string) error
}

type TokenGenerator interface {
	GenerateToken(userID user.UserID, username string) (string, error)
}

type LoginUserUseCase struct {
	userRepo       ports.UserRepository
	passwordHasher PasswordHasher
	tokenGen       TokenGenerator
}

func NewLoginUserUseCase(userRepo ports.UserRepository, hasher PasswordHasher, tokenGen TokenGenerator) *LoginUserUseCase {
	return &LoginUserUseCase{
		userRepo:       userRepo,
		passwordHasher: hasher,
		tokenGen:       tokenGen,
	}
}

func (uc *LoginUserUseCase) Execute(ctx context.Context, username, password string) (*user.User, string, error) {
	u, err := uc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, "", err 
	}
	if u == nil {
		return nil, "", ErrInvalidCredentials
	}

	if err := uc.passwordHasher.Compare(u.PasswordHash, password); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := uc.tokenGen.GenerateToken(u.ID, u.Username)
	if err != nil {
		return nil, "", err
	}

	return u, token, nil
}
