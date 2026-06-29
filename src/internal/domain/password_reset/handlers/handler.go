package passwordResetHandlers

import (
	pwdSvc "backend/src/internal/domain/password_reset/services"
	"backend/src/pkg/apperr"
	"backend/src/pkg/respond"

	"github.com/gin-gonic/gin"
)

type PasswordResetHandler struct {
	Svc *pwdSvc.PasswordResetService
}

func New(svc *pwdSvc.PasswordResetService) *PasswordResetHandler {
	return &PasswordResetHandler{Svc: svc}
}

// POST /auth/forgot-password
func (h *PasswordResetHandler) Request(c *gin.Context) {
	var body struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respond.Err(c, apperr.BadRequest("email inválido"))
		return
	}
	// Always returns OK to avoid email enumeration
	_ = h.Svc.RequestReset(body.Email)
	respond.OK(c, gin.H{"message": "se o email existir, você receberá um link em breve"})
}

// POST /auth/reset-password
func (h *PasswordResetHandler) Reset(c *gin.Context) {
	var body struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respond.Err(c, apperr.BadRequest(err.Error()))
		return
	}
	if err := h.Svc.Reset(body.Token, body.Password); err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, gin.H{"message": "senha atualizada com sucesso"})
}
