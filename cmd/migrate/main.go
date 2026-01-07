package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"image-processing-service/internal/config"
	"image-processing-service/internal/database"
)

func loadEnv() {
	file, err := os.Open(".env")
	if err != nil {
		log.Printf("Warning: .env file not found or readable: %v", err)
		return
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			if serr := os.Setenv(key, val); serr != nil {
				log.Printf("Warning: failed to set env %s: %v", key, serr)
			}
		}
	}
}

func main() {
	// 1. Load Env
	loadEnv()

	// 2. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 3. Connect to DB
	dbConfig, cerr := pgxpool.ParseConfig(cfg.Supabase.DBURL)
	if cerr != nil {
		log.Fatalf("Failed to parse DB config: %v", cerr)
	}

	pool, perr := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if perr != nil {
		log.Fatalf("Failed to connect to DB: %v", perr)
	}
	defer pool.Close()

	if p_err := pool.Ping(context.Background()); p_err != nil {
		log.Fatalf("Failed to ping DB: %v", p_err)
	}
	log.Println("Connected to Database successfully.")

	// 4. Execute Migrations
	ctx := context.Background()
	migrationDir := "migrations"
	if err := database.RunMigrations(ctx, pool, migrationDir); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("All migrations applied successfully.")
}
