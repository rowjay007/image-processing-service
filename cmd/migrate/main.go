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
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	for _, filename := range sqlFiles {
		log.Printf("Running migration: %s", filename)
		// G304 fix: Ensure the filename is just a filename and doesn't contain path traversal
		safePath := filepath.Join(migrationDir, filepath.Clean(filename))
		content, rerr := os.ReadFile(filepath.Clean(safePath))
		if rerr != nil {
			log.Fatalf("Failed to read file %s: %v", filename, rerr)
		}

		// Execute
		if _, txerr := tx.Exec(ctx, string(content)); txerr != nil {
			log.Fatalf("Failed to execute migration %s: %v", filename, txerr)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	log.Println("All migrations applied successfully.")
}
