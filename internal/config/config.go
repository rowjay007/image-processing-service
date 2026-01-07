package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
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
	v := viper.New()

	// Default values
	v.SetDefault("PORT", "8080")
	v.SetDefault("ENVIRONMENT", "development")
	v.SetDefault("GIN_MODE", "debug")

	v.SetDefault("DB_MAX_CONNS", 25)
	v.SetDefault("DB_MIN_CONNS", 5)
	v.SetDefault("DB_MAX_CONN_LIFETIME", time.Hour)
	v.SetDefault("DB_MAX_CONN_IDLE_TIME", 30*time.Minute)

	v.SetDefault("UPSTASH_REDIS_PORT", "6379")
	v.SetDefault("UPSTASH_REDIS_TLS", true)

	v.SetDefault("CLOUDINARY_FOLDER", "image-processing-service")
	v.SetDefault("CLOUDINARY_SECURE", true)
	v.SetDefault("CLOUDINARY_USE_AUTO_FORMAT", true)
	v.SetDefault("CLOUDINARY_USE_AUTO_QUALITY", true)

	v.SetDefault("QUEUE_NAME", "image-transform-jobs")
	v.SetDefault("QUEUE_DURABLE", true)
	v.SetDefault("QUEUE_PREFETCH_COUNT", 5)

	v.SetDefault("JWT_SECRET", "secret")
	v.SetDefault("JWT_EXPIRY", 24*time.Hour)
	v.SetDefault("JWT_ISSUER", "image-processing-service")

	v.SetDefault("MAX_UPLOAD_SIZE", 20971520)
	v.SetDefault("MAX_IMAGE_WIDTH", 8000)
	v.SetDefault("MAX_IMAGE_HEIGHT", 8000)
	v.SetDefault("RATE_LIMIT_UPLOADS", 100)
	v.SetDefault("RATE_LIMIT_TRANSFORMS", 500)
	v.SetDefault("RATE_LIMIT_WINDOW", time.Hour)

	// Environment mapping
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Load from .env if present
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	_ = v.ReadInConfig()

	return &Config{
		Server: ServerConfig{
			Port:        v.GetString("PORT"),
			Environment: v.GetString("ENVIRONMENT"),
			GinMode:     v.GetString("GIN_MODE"),
		},
		Supabase: SupabaseConfig{
			DBURL:           v.GetString("SUPABASE_DB_URL"),
			MaxConns:        v.GetInt("DB_MAX_CONNS"),
			MinConns:        v.GetInt("DB_MIN_CONNS"),
			MaxConnLifetime: v.GetDuration("DB_MAX_CONN_LIFETIME"),
			MaxConnIdleTime: v.GetDuration("DB_MAX_CONN_IDLE_TIME"),
		},
		Upstash: UpstashConfig{
			Host:     v.GetString("UPSTASH_REDIS_HOST"),
			Port:     v.GetString("UPSTASH_REDIS_PORT"),
			Password: v.GetString("UPSTASH_REDIS_PASSWORD"),
			TLS:      v.GetBool("UPSTASH_REDIS_TLS"),
		},
		Cloudinary: CloudinaryConfig{
			CloudName:      v.GetString("CLOUDINARY_CLOUD_NAME"),
			APIKey:         v.GetString("CLOUDINARY_API_KEY"),
			APISecret:      v.GetString("CLOUDINARY_API_SECRET"),
			Folder:         v.GetString("CLOUDINARY_FOLDER"),
			Secure:         v.GetBool("CLOUDINARY_SECURE"),
			UseAutoFormat:  v.GetBool("CLOUDINARY_USE_AUTO_FORMAT"),
			UseAutoQuality: v.GetBool("CLOUDINARY_USE_AUTO_QUALITY"),
		},
		CloudAMQP: CloudAMQPConfig{
			URL:           v.GetString("CLOUDAMQP_URL"),
			QueueName:     v.GetString("QUEUE_NAME"),
			QueueDurable:  v.GetBool("QUEUE_DURABLE"),
			PrefetchCount: v.GetInt("QUEUE_PREFETCH_COUNT"),
		},
		JWT: JWTConfig{
			Secret: v.GetString("JWT_SECRET"),
			Expiry: v.GetDuration("JWT_EXPIRY"),
			Issuer: v.GetString("JWT_ISSUER"),
		},
		Limits: LimitsConfig{
			MaxUploadSize:       v.GetInt64("MAX_UPLOAD_SIZE"),
			MaxImageWidth:       v.GetInt("MAX_IMAGE_WIDTH"),
			MaxImageHeight:      v.GetInt("MAX_IMAGE_HEIGHT"),
			RateLimitUploads:    v.GetInt("RATE_LIMIT_UPLOADS"),
			RateLimitTransforms: v.GetInt("RATE_LIMIT_TRANSFORMS"),
			RateLimitWindow:     v.GetDuration("RATE_LIMIT_WINDOW"),
		},
	}, nil
}
