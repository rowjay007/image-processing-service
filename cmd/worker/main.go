package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"image-processing-service/internal/adapters/logging"
	"image-processing-service/internal/adapters/persistence"
	"image-processing-service/internal/adapters/processor"
	"image-processing-service/internal/adapters/queue"
	"image-processing-service/internal/adapters/storage"
	"image-processing-service/internal/config"
	"image-processing-service/internal/ports"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// 2. Init Logger
	logger, err := logging.NewLogger(cfg.Server.Environment)
	if err != nil {
		panic("Failed to init logger: " + err.Error())
	}
	defer func() {
		_ = logger.Sync()
	}()
	zap.ReplaceGlobals(logger)

	// 3. Dependencies
	// DB
	dbConfig, err := pgxpool.ParseConfig(cfg.Supabase.DBURL)
	if err != nil {
		logger.Fatal("Failed to parse DB config", zap.Error(err))
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		logger.Fatal("Failed to connect to DB", zap.Error(err))
	}
	defer pool.Close()

	_ = persistence.NewPostgresImageRepository(pool)

	// Storage
	_, err = storage.NewCloudinaryStorage(cfg.Cloudinary)
	if err != nil {
		logger.Fatal("Failed to init storage", zap.Error(err))
	}

	// Queue
	q, err := queue.NewCloudAMQPQueue(cfg.CloudAMQP)
	if err != nil {
		logger.Fatal("Failed to init queue", zap.Error(err))
	}
	defer func() {
		_ = q.Close()
	}()

	// Processor
	imgProcessor := processor.NewStdLibImageProcessor()

	// 4. Consumer Logic
	logger.Info("Worker starting...")

	err = q.Consume(context.Background(), func(job *ports.TransformJob) error {
		logger.Info("Processing Job",
			zap.String("job_id", job.JobID),
			zap.String("image_id", job.ImageID),
		)

		_, terr := imgProcessor.Transform(context.Background(), nil, job.Spec)
		if terr != nil {
			logger.Warn("Processor Error (expected usage of bimg)", zap.Error(terr))
			return nil
		}

		return nil
	})

	if err != nil {
		logger.Fatal("Failed to start consumer", zap.Error(err))
	}

	// Wait for signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Worker shutting down...")
}
