package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"image-processing-service/internal/adapters/http/middleware"
	"image-processing-service/internal/container"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize Container
	c, err := container.NewContainer()
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}
	defer c.Close()

	// Setup Router
	if c.Config.Server.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Middlewares
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.SecurityHeadersMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now()})
	})

	// API Routes
	v1 := r.Group("/api/v1")
	{
		// Auth Routes
		auth := v1.Group("/auth")
		auth.Use(c.RateLimitMiddleware.Limit(5, 15*time.Minute))
		{
			auth.POST("/register", c.AuthHandler.Register)
			auth.POST("/login", c.AuthHandler.Login)
		}

		// Protected Routes
		protected := v1.Group("/")
		protected.Use(c.AuthMiddleware.Handle())
		{
			protected.GET("/me", func(ctx *gin.Context) {
				userID, _ := ctx.Get("userID")
				username, _ := ctx.Get("username")
				ctx.JSON(http.StatusOK, gin.H{
					"user_id":  userID,
					"username": username,
					"message":  "You are authenticated!",
				})
			})

			// Image Routes
			images := protected.Group("/images")
			{
				images.POST("", c.RateLimitMiddleware.Limit(10, time.Hour), c.ImageHandler.Upload)
				images.POST("/:id/transform", c.RateLimitMiddleware.Limit(50, time.Minute), c.ImageHandler.Transform)
				images.GET("", c.ImageHandler.List)
				images.GET("/:id", c.ImageHandler.Get)
			}
		}
	}

	// Start Server
	srv := &http.Server{
		Addr:    ":" + c.Config.Server.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on port %s", c.Config.Server.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
