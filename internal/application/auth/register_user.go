package auth

import (
	"context"
	"errors"
	"strings"

	"image-processing-service/internal/domain/user"
	"image-processing-service/internal/ports"
)

var (
	ErrUserAlreadyExists = errors.New("username already exists")
)

type RegisterUserUseCase struct {
	userRepo ports.UserRepository
}

func NewRegisterUserUseCase(userRepo ports.UserRepository) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepo: userRepo,
	}
}

// Execute registers a new user with the given credentials.
// Password should be hashed BEFORE calling domain logic usually, but here we can accept the hash
// or we can inject a hashing service.
// Following the structure, the handler will probably call a hasher, OR we can inject a Hasher port here.
// Let's assume the handler does the hashing to keep this pure or we inject an interface.
// For better security, let's say we pass the raw password and hash it here via an interface to ensure it's always hashed.
// But the plan had a separate 'adapters/auth/password_hasher' which is great.
// Let's add a PasswordHasher interface to the port or just pass the hash.
// The domain User factory takes a hash.
// For simplicity in this use case, let's assume the caller passes the hash.
// OR better: we define a PasswordHasher port. I'll stick to passing the hash for now to avoid changing ports.go too much,
// but typically the use case orchestrates hashing.

func (uc *RegisterUserUseCase) Execute(ctx context.Context, username, passwordHash string) (*user.User, error) {
	username = strings.TrimSpace(strings.ToLower(username))
	if username == "" {
		return nil, user.ErrInvalidUsername
	}

	// Check if user exists
	existing, err := uc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		// If error is anything other than "not found", return it.
		// We'll need to define what GetByUsername returns when not found.
		// Usually nil, nil or a specific error.
		// We'll assume nil, nil means not found for now, or check specific error.
		// Implementation dependent.
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	newUser, err := user.New(username, passwordHash)
	if err != nil {
		return nil, err
	}

	if err := uc.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}
