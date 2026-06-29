package middleware

import (
	"backend/src/internal/repository/user"
	"backend/src/utils"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var authSvc = utils.NewAuthService()

type MiddleWare struct {
	UserRepo *user.UserRepository
}

func NewMiddleWare(userRepo *user.UserRepository) *MiddleWare {
	return &MiddleWare{UserRepo: userRepo}
}

func (m *MiddleWare) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := authSvc.ValidateToken(tokenString)
		if err != nil {
			if errors.Is(err, utils.ErrTokenExpired) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			}
			c.Abort()
			return
		}

		exists, err := m.UserRepo.GetUserByID(claims.UserID)
		if err != nil || exists == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
