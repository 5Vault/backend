package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateMiddleware struct {
	RedisClient *redis.Client
}

func NewRateMiddleware(redisClient *redis.Client) *RateMiddleware {
	return &RateMiddleware{
		RedisClient: redisClient,
	}
}

func (r *RateMiddleware) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP address
		clientIP := c.ClientIP()

		if clientIP == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to determine client IP"})
			c.Abort()
			return
		}

		if r.RedisClient == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Redis client not initialized"})
			c.Abort()
			return
		}

		ctx := context.Background()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		// Increment the counter
		count, err := r.RedisClient.Incr(ctx, key).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit check failed"})
			c.Abort()
			return
		}

		// Set expiration of 1 second on first request (when count is 1)
		if count == 1 {
			r.RedisClient.Expire(ctx, key, 2*time.Second)
		}

		// Check if limit exceeded
		if count > 5 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
