package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type KeyMiddleware struct {
	publicKey  string
	privateKey string
}

func NewKeyMiddleware() *KeyMiddleware {
	return &KeyMiddleware{
		publicKey:  os.Getenv("PUBLIC_KEY"),
		privateKey: os.Getenv("PRIVATE_KEY"),
	}
}

// ValidateKeysMiddleware verifica se as chaves Public-Key e Private-Key estão presentes e válidas nos headers
func (k *KeyMiddleware) ValidateKeysMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		publicKey := c.GetHeader("Public-Key")
		privateKey := c.GetHeader("Private-Key")

		// Verifica se os headers estão presentes
		if publicKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Public-Key header is required"})
			c.Abort()
			return
		}

		if privateKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Private-Key header is required"})
			c.Abort()
			return
		}

		// Remove espaços em branco
		publicKey = strings.TrimSpace(publicKey)
		privateKey = strings.TrimSpace(privateKey)

		// Valida as chaves contra os valores esperados
		if publicKey != k.publicKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Public-Key"})
			c.Abort()
			return
		}

		if privateKey != k.privateKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Private-Key"})
			c.Abort()
			return
		}

		// Define as chaves no contexto para uso posterior
		c.Set("public_key", publicKey)
		c.Set("private_key", privateKey)

		c.Next()
	}
}

// ValidatePublicKeyOnly valida apenas a chave pública (para endpoints menos sensíveis)
func (k *KeyMiddleware) ValidatePublicKeyOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		publicKey := c.GetHeader("Public-Key")

		if publicKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Public-Key header is required"})
			c.Abort()
			return
		}

		publicKey = strings.TrimSpace(publicKey)

		if publicKey != k.publicKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Public-Key"})
			c.Abort()
			return
		}

		c.Set("public_key", publicKey)
		c.Next()
	}
}
