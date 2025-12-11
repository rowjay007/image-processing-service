package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// UserID is a strongly typed identifier for a user.
type UserID string

// User represents a registered user in the system.
type User struct {
	ID           UserID
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

var (
	ErrInvalidUsername = errors.New("invalid username")
	ErrInvalidPassword = errors.New("invalid password")
)

// New creates a new User instance.
func New(username, passwordHash string) (*User, error) {
	if username == "" {
		return nil, ErrInvalidUsername
	}
	if passwordHash == "" {
		return nil, ErrInvalidPassword
	}

	return &User{
		ID:           UserID(uuid.New().String()),
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	}, nil
}
