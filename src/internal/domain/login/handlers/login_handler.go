package lHandlers

import (
	lgServices "backend/src/internal/domain/login/services"
	"backend/src/internal/models"
	"backend/src/pkg/apperr"
	"backend/src/pkg/respond"

	"github.com/gin-gonic/gin"
)

type LoginHandler struct {
	LoginService *lgServices.LoginService
}

func NewLoginHandler(service *lgServices.LoginService) *LoginHandler {
	return &LoginHandler{LoginService: service}
}

func (l *LoginHandler) Try(c *gin.Context) {
	var req models.RequestLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Err(c, apperr.BadRequest("username and password are required"))
		return
	}
	token, err := l.LoginService.Try(&req, c.ClientIP())
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, gin.H{"token": *token})
}
