package middleware

import (
	"backend/src/utils"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var rateAuthSvc = utils.NewAuthService()

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
		// Skip rate limiting for specific endpoints
		if c.Request.Method == "GET" && c.FullPath() == "/api/v1/file/:id" {
			c.Next()
			return
		}

		if r.RedisClient == nil {
			c.Next()
			return
		}

		// Try to identify user by JWT token or API key first, then fallback to IP
		identity := c.GetString("user_id")
		if identity == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString := strings.TrimPrefix(authHeader, "Bearer ")
				claims, err := rateAuthSvc.ValidateToken(tokenString)
				if err == nil && claims != nil {
					identity = claims.UserID
				}
			}
		}

		if identity == "" {
			apiKey := c.GetHeader("Api-Key")
			if apiKey != "" {
				identity = "apikey:" + apiKey
			}
		}

		if identity == "" {
			identity = c.ClientIP()
		}

		ctx := context.Background()
		key := fmt.Sprintf("rate_limit:%s", identity)

		count, err := r.RedisClient.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		// Ensure TTL is set
		if count == 1 {
			r.RedisClient.Expire(ctx, key, 10*time.Second)
		} else {
			// Prevent stuck keys without TTL due to race conditions
			ttl, err := r.RedisClient.TTL(ctx, key).Result()
			if err == nil && ttl < 0 {
				r.RedisClient.Expire(ctx, key, 10*time.Second)
			}
		}

		// 200 requests per 10 seconds per user
		if count > 200 {
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
