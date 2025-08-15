package middleware

import (
	lServices "backend/src/internal/domain/login/services"
	"backend/src/internal/domain/user/repositories"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var lSvcs = lServices.NewAuthService()

type MiddleWare struct {
	UserRepo *repositories.UserRepository
}

func NewMiddleWare(userRepo *repositories.UserRepository) *MiddleWare {
	return &MiddleWare{
		UserRepo: userRepo,
	}
}

// AuthMiddleware is a middleware function that checks if the user is authenticated
func (m *MiddleWare) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header not provided"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := lSvcs.ValidateToken(tokenString)

		user, err := m.UserRepo.GetUserByID(claims.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
