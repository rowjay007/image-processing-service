package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"image-processing-service/internal/config"
)

func loadEnv() {
	file, err := os.Open(".env")
	if err != nil {
		log.Printf("Warning: .env file not found or readable: %v", err)
		return
	}
	defer file.Close()

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
			os.Setenv(key, val)
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
	dbConfig, err := pgxpool.ParseConfig(cfg.Supabase.DBURL)
	if err != nil {
		log.Fatalf("Failed to parse DB config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}
	log.Println("Connected to Database successfully.")

	// 4. Find Migration Files
	migrationDir := "migrations"
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		log.Fatalf("Failed to read migration directory: %v", err)
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	
	// Sort by validation (ensure 001 runs before 002)
	sort.Strings(sqlFiles)

	// 5. Execute Migrations
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	for _, filename := range sqlFiles {
		log.Printf("Running migration: %s", filename)
		content, err := os.ReadFile(filepath.Join(migrationDir, filename))
		if err != nil {
			log.Fatalf("Failed to read file %s: %v", filename, err)
		}

		// Execute
		if _, err := tx.Exec(ctx, string(content)); err != nil {
			log.Fatalf("Failed to execute migration %s: %v", filename, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	log.Println("All migrations applied successfully.")
}
