package clients

import (
	"context"
	"fmt"

	authv1 "image-processing-service/api/gen/go/auth/v1"
	"image-processing-service/internal/ports"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthGRPCClient struct {
	client authv1.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthGRPCClient(addr string) (*AuthGRPCClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	return &AuthGRPCClient{
		client: authv1.NewAuthServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *AuthGRPCClient) ValidateToken(token string) (*ports.Claims, error) {
	resp, err := c.client.ValidateToken(context.Background(), &authv1.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		return nil, err
	}

	if !resp.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return &ports.Claims{
		UserID:   resp.UserId,
		Username: resp.Username,
	}, nil
}

func (c *AuthGRPCClient) Close() error {
	return c.conn.Close()
}

// We can also implement Register/Login if we want the API to proxy these calls
func (c *AuthGRPCClient) Register(ctx context.Context, username, password string) (string, string, error) {
	resp, err := c.client.Register(ctx, &authv1.RegisterRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", "", err
	}
	return resp.UserId, resp.Username, nil
}

func (c *AuthGRPCClient) Login(ctx context.Context, username, password string) (string, string, string, error) {
	resp, err := c.client.Login(ctx, &authv1.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", "", "", err
	}
	return resp.UserId, resp.Username, resp.Token, nil
}
