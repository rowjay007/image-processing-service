package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server     ServerConfig
	Supabase   SupabaseConfig
	Upstash    UpstashConfig
	Cloudinary CloudinaryConfig
	CloudAMQP  CloudAMQPConfig
	JWT        JWTConfig
	Limits     LimitsConfig
}

type ServerConfig struct {
	Port        string
	Environment string
	GinMode     string
}

type SupabaseConfig struct {
	DBURL           string
	MaxConns        int
	MinConns        int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

type UpstashConfig struct {
	Host     string
	Port     string
	Password string
	TLS      bool
}

type CloudinaryConfig struct {
	CloudName      string
	APIKey         string
	APISecret      string
	Folder         string
	Secure         bool
	UseAutoFormat  bool
	UseAutoQuality bool
}

type CloudAMQPConfig struct {
	URL           string
	QueueName     string
	QueueDurable  bool
	PrefetchCount int
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
	Issuer string
}

type LimitsConfig struct {
	MaxUploadSize       int64
	MaxImageWidth       int
	MaxImageHeight      int
	RateLimitUploads    int
	RateLimitTransforms int
	RateLimitWindow     time.Duration
}

func LoadConfig() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port:        getEnv("PORT", "8080"),
			Environment: getEnv("ENVIRONMENT", "development"),
			GinMode:     getEnv("GIN_MODE", "debug"),
		},
		Supabase: SupabaseConfig{
			DBURL:           getEnv("SUPABASE_DB_URL", ""),
			MaxConns:        getEnvInt("DB_MAX_CONNS", 25),
			MinConns:        getEnvInt("DB_MIN_CONNS", 5),
			MaxConnLifetime: getEnvDuration("DB_MAX_CONN_LIFETIME", time.Hour),
			MaxConnIdleTime: getEnvDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
		},
		Upstash: UpstashConfig{
			Host:     getEnv("UPSTASH_REDIS_HOST", ""),
			Port:     getEnv("UPSTASH_REDIS_PORT", "6380"),
			Password: getEnv("UPSTASH_REDIS_PASSWORD", ""),
			TLS:      getEnvBool("UPSTASH_REDIS_TLS", true),
		},
		Cloudinary: CloudinaryConfig{
			CloudName:      getEnv("CLOUDINARY_CLOUD_NAME", ""),
			APIKey:         getEnv("CLOUDINARY_API_KEY", ""),
			APISecret:      getEnv("CLOUDINARY_API_SECRET", ""),
			Folder:         getEnv("CLOUDINARY_FOLDER", "image-processing-service"),
			Secure:         getEnvBool("CLOUDINARY_SECURE", true),
			UseAutoFormat:  getEnvBool("CLOUDINARY_USE_AUTO_FORMAT", true),
			UseAutoQuality: getEnvBool("CLOUDINARY_USE_AUTO_QUALITY", true),
		},
		CloudAMQP: CloudAMQPConfig{
			URL:           getEnv("CLOUDAMQP_URL", ""),
			QueueName:     getEnv("QUEUE_NAME", "image-transform-jobs"),
			QueueDurable:  getEnvBool("QUEUE_DURABLE", true),
			PrefetchCount: getEnvInt("QUEUE_PREFETCH_COUNT", 5),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "secret"),
			Expiry: getEnvDuration("JWT_EXPIRY", 15*time.Minute),
			Issuer: getEnv("JWT_ISSUER", "image-processing-service"),
		},
		Limits: LimitsConfig{
			MaxUploadSize:       getEnvInt64("MAX_UPLOAD_SIZE", 20971520),
			MaxImageWidth:       getEnvInt("MAX_IMAGE_WIDTH", 8000),
			MaxImageHeight:      getEnvInt("MAX_IMAGE_HEIGHT", 8000),
			RateLimitUploads:    getEnvInt("RATE_LIMIT_UPLOADS", 100),
			RateLimitTransforms: getEnvInt("RATE_LIMIT_TRANSFORMS", 500),
			RateLimitWindow:     getEnvDuration("RATE_LIMIT_WINDOW", time.Hour),
		},
	}, nil
}

// Helpers

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if durationVal, err := time.ParseDuration(value); err == nil {
			return durationVal
		}
		// Try parsing as integer seconds if duration parse fails
		if intVal, err := strconv.Atoi(value); err == nil {
			return time.Duration(intVal) * time.Second
		}
		log.Printf("Warning: Invalid duration format for %s: %s. Using fallback.", key, value)
	}
	return fallback
}
