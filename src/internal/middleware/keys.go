package middleware

import (
	"backend/src/internal/domain/key/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

type KeyMiddleware struct {
	KeyRepo *repository.KeyRepository
}

func NewKeyMiddleware(repo *repository.KeyRepository) *KeyMiddleware {
	return &KeyMiddleware{
		KeyRepo: repo,
	}
}

// ValidateKeysMiddleware verifica se as chaves Public-Key e Private-Key estão presentes e válidas nos headers
func (k *KeyMiddleware) ValidateKeysMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("Api-Key")

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Api-Key header is required"})
		}

		result, err := k.KeyRepo.GetKey(apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Api-Key"})
			c.Abort()
			return
		}
		if result == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Api-Key not found"})
			c.Abort()
			return
		}
		c.Set("api_key", result.Key)
		c.Next()
	}
}
