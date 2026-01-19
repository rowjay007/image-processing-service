package handlers

import (
	"context"

	authv1 "image-processing-service/api/gen/go/auth/v1"
	"image-processing-service/internal/application/auth"
	"image-processing-service/internal/ports"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthGRPCHandler struct {
	authv1.UnimplementedAuthServiceServer
	registerUC *auth.RegisterUserUseCase
	loginUC    *auth.LoginUserUseCase
	hasher     auth.PasswordHasher
	jwtProv    ports.AuthProvider
}

func NewAuthGRPCHandler(
	registerUC *auth.RegisterUserUseCase,
	loginUC *auth.LoginUserUseCase,
	hasher auth.PasswordHasher,
	jwtProv ports.AuthProvider,
) *AuthGRPCHandler {
	return &AuthGRPCHandler{
		registerUC: registerUC,
		loginUC:    loginUC,
		hasher:     hasher,
		jwtProv:    jwtProv,
	}
}

func (h *AuthGRPCHandler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	// Simple hashing for now (ideally hashed in use case if we add a hasher port there)
	// For now, mirroring the HTTP handler logic
	// But actually, we should probably hash it here or ensure the use case expects raw + hasher.
	// Looking at current HTTP handler: it hashes before calling UC.
	// Let's stick to that for parity, but in a microservice we might want to move hashing into UC.
	
	hash, err := h.hasher.Hash(req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	u, err := h.registerUC.Execute(ctx, req.Username, hash)
	if err != nil {
		if err == auth.ErrUserAlreadyExists {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "registration failed: %v", err)
	}

	return &authv1.RegisterResponse{
		UserId:   string(u.ID),
		Username: u.Username,
	}, nil
}

func (h *AuthGRPCHandler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	u, token, err := h.loginUC.Execute(ctx, req.Username, req.Password)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "login failed: %v", err)
	}

	return &authv1.LoginResponse{
		Token:    token,
		UserId:   string(u.ID),
		Username: u.Username,
	}, nil
}

func (h *AuthGRPCHandler) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	claims, err := h.jwtProv.ValidateToken(req.Token)
	if err != nil {
		return &authv1.ValidateTokenResponse{Valid: false}, nil
	}

	return &authv1.ValidateTokenResponse{
		Valid:    true,
		UserId:   claims.UserID,
		Username: claims.Username,
	}, nil
}
