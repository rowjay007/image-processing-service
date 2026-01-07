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

	logger, lerr := logging.NewLogger(cfg.Server.Environment)
	if lerr != nil {
		return nil, fmt.Errorf("failed to init logger: %w", lerr)
	}
	zap.ReplaceGlobals(logger)

	dbConfig, dberr := pgxpool.ParseConfig(cfg.Supabase.DBURL)
	if dberr != nil {
		return nil, fmt.Errorf("failed to parse db config: %w", dberr)
	}
	// #nosec G115
	dbConfig.MaxConns = int32(cfg.Supabase.MaxConns)
	// #nosec G115
	dbConfig.MinConns = int32(cfg.Supabase.MinConns)
	dbConfig.MaxConnLifetime = cfg.Supabase.MaxConnLifetime
	dbConfig.MaxConnIdleTime = cfg.Supabase.MaxConnIdleTime

	pool, p_err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if p_err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", p_err)
	}

	if perr := pool.Ping(context.Background()); perr != nil {
		log.Printf("Warning: Failed to ping database: %v", perr)
	}

	userRepo := persistence.NewPostgresUserRepository(pool)
	imageRepo := persistence.NewPostgresImageRepository(pool)

	storage, serr := storage.NewCloudinaryStorage(cfg.Cloudinary)
	if serr != nil {
		return nil, fmt.Errorf("failed to init storage: %w", serr)
	}

	imgProcessor := processor.NewBimgProcessor()

	q, qerr := queue.NewCloudAMQPQueue(cfg.CloudAMQP)
	if qerr != nil {
		return nil, fmt.Errorf("failed to init queue: %w", qerr)
	}

	// Redis Cache & Rate Limiting (Optional/Resilient)
	var cacheSvc ports.Cache
	var rateLimiter ports.RateLimiter

	redisSvc, rerr := cache.NewRedisCache(cfg.Upstash)
	if rerr != nil {
		log.Printf("Warning: Failed to init redis cache: %v. Caching and Rate Limiting will be DISABLED.", rerr)
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

	uploadUC := appImage.NewUploadImageUseCase(imageRepo, storage, imgProcessor)
	asyncTransformUC := appImage.NewAsyncTransformImageUseCase(imageRepo, q)
	syncTransformUC := appImage.NewTransformImageSyncUseCase(imageRepo, storage, imgProcessor)
	getUC := appImage.NewGetImageUseCase(imageRepo, cacheSvc)
	listUC := appImage.NewListImagesUseCase(imageRepo, cacheSvc)

	authHandler := handlers.NewAuthHandler(registerUC, loginUC, hasher)
	authMiddleware := middleware.NewAuthMiddleware(jwtProvider)
	imageHandler := handlers.NewImageHandler(uploadUC, asyncTransformUC, syncTransformUC, getUC, listUC)

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
