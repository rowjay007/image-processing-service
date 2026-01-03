package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"image-processing-service/internal/ports"
)

type RateLimitMiddleware struct {
	limiter ports.RateLimiter
}

func NewRateLimitMiddleware(limiter ports.RateLimiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{limiter: limiter}
}

func (m *RateLimitMiddleware) Limit(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if sub, exists := c.Get("userID"); exists {
			key = fmt.Sprintf("%v", sub)
		}

		key = fmt.Sprintf("rl:%s:%s", c.FullPath(), key)

		allowed, err := m.limiter.Allow(c.Request.Context(), key, limit, window)
		if err != nil {
			c.Next()
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}
