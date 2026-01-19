package ports

import (
	"context"
)

type RemoteAuthService interface {
	Register(ctx context.Context, username, password string) (string, string, error)
	Login(ctx context.Context, username, password string) (string, string, string, error)
	ValidateToken(token string) (*Claims, error)
}
