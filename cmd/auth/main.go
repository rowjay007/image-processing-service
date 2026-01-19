package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	authv1 "image-processing-service/api/gen/go/auth/v1"
	"image-processing-service/internal/adapters/auth"
	"image-processing-service/internal/adapters/grpc/handlers"
	"image-processing-service/internal/adapters/persistence"
	appAuth "image-processing-service/internal/application/auth"
	"image-processing-service/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Database connection
	dbPool, err := pgxpool.New(context.Background(), cfg.Supabase.DBURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Repositories
	userRepo := persistence.NewPostgresUserRepository(dbPool)

	// Auth adapters
	jwtProvider := auth.NewJWTProvider(cfg.JWT)
	hasher := auth.NewBcryptPasswordHasher()

	// Use cases
	registerUC := appAuth.NewRegisterUserUseCase(userRepo)
	loginUC := appAuth.NewLoginUserUseCase(userRepo, hasher, jwtProvider)

	// gRPC Handler
	authHandler := handlers.NewAuthGRPCHandler(registerUC, loginUC, hasher, jwtProvider)

	// gRPC Server setup
	grpcServer := grpc.NewServer()
	authv1.RegisterAuthServiceServer(grpcServer, authHandler)
	reflection.Register(grpcServer)

	// Listen
	port := os.Getenv("AUTH_SERVICE_PORT")
	if port == "" {
		port = "50051"
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Graceful shutdown
	go func() {
		log.Printf("Auth Service starting on port %s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down Auth Service...")
	grpcServer.GracefulStop()
}
