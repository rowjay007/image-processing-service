package container

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"image-processing-service/internal/adapters/auth"
	"image-processing-service/internal/adapters/cache"
	"image-processing-service/internal/adapters/http/handlers"
	"image-processing-service/internal/adapters/http/middleware"
	"image-processing-service/internal/adapters/logging"
	"image-processing-service/internal/adapters/persistence"
	"image-processing-service/internal/adapters/processor"
	"image-processing-service/internal/adapters/queue"
	"image-processing-service/internal/adapters/storage"
	appAuth "image-processing-service/internal/application/auth"
	appImage "image-processing-service/internal/application/image"
	"image-processing-service/internal/config"
	"image-processing-service/internal/ports"
)

type Container struct {
	Config *config.Config
	Logger *zap.Logger
	DB     *pgxpool.Pool

	AuthHandler    *handlers.AuthHandler
	AuthMiddleware *middleware.AuthMiddleware

	ImageHandler *handlers.ImageHandler

	RateLimitMiddleware *middleware.RateLimitMiddleware
}

func NewContainer() (*Container, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	logger, err := logging.NewLogger(cfg.Server.Environment)
	if err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}
	zap.ReplaceGlobals(logger)

	dbConfig, err := pgxpool.ParseConfig(cfg.Supabase.DBURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse db config: %w", err)
	}
	dbConfig.MaxConns = int32(cfg.Supabase.MaxConns)
	dbConfig.MinConns = int32(cfg.Supabase.MinConns)
	dbConfig.MaxConnLifetime = cfg.Supabase.MaxConnLifetime
	dbConfig.MaxConnIdleTime = cfg.Supabase.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	
	if err := pool.Ping(context.Background()); err != nil {
		log.Printf("Warning: Failed to ping database: %v", err)
	}

	userRepo := persistence.NewPostgresUserRepository(pool)
	imageRepo := persistence.NewPostgresImageRepository(pool)
	
	storage, err := storage.NewCloudinaryStorage(cfg.Cloudinary)
	if err != nil {
		return nil, fmt.Errorf("failed to init storage: %w", err)
	}
	
	processor := processor.NewStdLibImageProcessor()
	
	queue, err := queue.NewCloudAMQPQueue(cfg.CloudAMQP)
	if err != nil {
		return nil, fmt.Errorf("failed to init queue: %w", err)
	}
	
	// Redis Cache & Rate Limiting (Optional/Resilient)
	var cacheSvc ports.Cache
	var rateLimiter ports.RateLimiter

	redisSvc, err := cache.NewRedisCache(cfg.Upstash)
	if err != nil {
		log.Printf("Warning: Failed to init redis cache: %v. Caching and Rate Limiting will be DISABLED.", err)
		cacheSvc = cache.NewNoOpCache()
		rateLimiter = cache.NewNoOpRateLimiter()
	} else {
		cacheSvc = redisSvc
		rateLimiter = cache.NewRedisRateLimiter(redisSvc.Client())
	}

	jwtProvider := auth.NewJWTProvider(cfg.JWT)
	hasher := auth.NewBcryptPasswordHasher()

	registerUC := appAuth.NewRegisterUserUseCase(userRepo)
	loginUC := appAuth.NewLoginUserUseCase(userRepo, hasher, jwtProvider)
	
	uploadUC := appImage.NewUploadImageUseCase(imageRepo, storage, processor)
	asyncTransformUC := appImage.NewAsyncTransformImageUseCase(imageRepo, queue)
	getUC := appImage.NewGetImageUseCase(imageRepo, cacheSvc)
	listUC := appImage.NewListImagesUseCase(imageRepo, cacheSvc)

	authHandler := handlers.NewAuthHandler(registerUC, loginUC, hasher)
	authMiddleware := middleware.NewAuthMiddleware(jwtProvider)
	imageHandler := handlers.NewImageHandler(uploadUC, asyncTransformUC, getUC, listUC)

	rateLimitMiddleware := middleware.NewRateLimitMiddleware(rateLimiter)

	return &Container{
		Config:              cfg,
		Logger:              logger,
		DB:                  pool,
		AuthHandler:         authHandler,
		AuthMiddleware:      authMiddleware,
		ImageHandler:        imageHandler,
		RateLimitMiddleware: rateLimitMiddleware,
	}, nil
}

func (c *Container) Close() {
	if c.DB != nil {
		c.DB.Close()
	}
}
