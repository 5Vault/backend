package middleware

import (
	"backend/src/internal/repository/user"
	lServices "backend/src/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var lSvcs = lServices.NewAuthService()

type MiddleWare struct {
	UserRepo *user.UserRepository
}

func NewMiddleWare(userRepo *user.UserRepository) *MiddleWare {
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

		consultUser, err := m.UserRepo.GetUserByID(claims.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		if consultUser == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
