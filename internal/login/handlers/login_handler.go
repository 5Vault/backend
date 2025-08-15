package lHandlers

import (
	lgServices "backend/internal/login/services"
	"backend/internal/models"

	"github.com/gin-gonic/gin"
)

type LoginHandler struct {
	LoginService *lgServices.LoginService
}

func NewLoginHandler(LoginService *lgServices.LoginService) *LoginHandler {
	return &LoginHandler{
		LoginService: LoginService,
	}
}

func (l *LoginHandler) Try(c *gin.Context) {
	var request *models.RequestLogin
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	token, err := l.LoginService.Try(request)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"token": *token})
}
