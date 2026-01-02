package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"image-processing-service/internal/adapters/persistence"
	"image-processing-service/internal/adapters/processor"
	"image-processing-service/internal/adapters/queue"
	"image-processing-service/internal/adapters/storage"
	"image-processing-service/internal/config"
	"image-processing-service/internal/ports"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Dependencies
	// DB
	dbConfig, err := pgxpool.ParseConfig(cfg.Supabase.DBURL)
	if err != nil {
		log.Fatalf("Failed to parse DB config: %v", err)
	}
	
	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer pool.Close()

	// Logic: Worker needs Repo to save variant status (if we implemented that)
	// But our consumer currently only Logs.
	// We keep Repo available for future logic.
	_ = persistence.NewPostgresImageRepository(pool)

	// Storage
	// Worker needs storage to download original and upload variant
	_ , err = storage.NewCloudinaryStorage(cfg.Cloudinary)
	if err != nil {
		log.Fatalf("Failed to init storage: %v", err)
	}

	// Queue
	q, err := queue.NewCloudAMQPQueue(cfg.CloudAMQP)
	if err != nil {
		log.Fatalf("Failed to init queue: %v", err)
	}
	defer q.Close()

	// Processor
	imgProcessor := processor.NewStdLibImageProcessor()

	// 3. Consumer Logic
	log.Println("Worker starting...")
	
	err = q.Consume(context.Background(), func(job *ports.TransformJob) error {
		log.Printf("Processing Job: %s for Image: %s", job.JobID, job.ImageID)
		
		// Logic would go here.
		// Since StdLib processor fails, we just log.
		
		// Fix: job.Spec is already *TransformationSpec (pointer) in the new struct definition
		_, err := imgProcessor.Transform(context.Background(), nil, job.Spec)
		if err != nil {
			log.Printf("Processor Error (expected usage of bimg): %v", err)
			return nil
		}
		
		return nil
	})
	
	if err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}

	// Wait for signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Worker shutting down...")
}
