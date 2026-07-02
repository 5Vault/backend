package middleware

import (
	"backend/src/internal/repository/key"
	"net/http"

	"github.com/gin-gonic/gin"
)

type KeyMiddleware struct {
	KeyRepo *key.KeyRepository
}

func NewKeyMiddleware(repo *key.KeyRepository) *KeyMiddleware {
	return &KeyMiddleware{
		KeyRepo: repo,
	}
}

func (k *KeyMiddleware) ValidateKeysMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("Api-Key")

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Api-Key header is required"})
			c.Abort()
			return
		}

		result, err := k.KeyRepo.GetKey(apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"valid_key": false})
			c.Abort()
			return
		}
		if result == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Api-Key not found"})
			c.Abort()
			return
		}

		c.Set("user_id_key", result.UserID)
		c.Set("api_key", result.Key)
		c.Next()
	}
}
