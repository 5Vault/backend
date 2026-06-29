package handlers

import (
	"backend/src/pkg/apperr"
	"backend/src/pkg/respond"
	"bytes"
	"encoding/base64"
	"image/png"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
)

// POST /user/2fa/setup — gera secret e retorna QR code
func (h *UserHandler) Setup2FA(c *gin.Context) {
	userID := c.GetString("user_id")
	user, err := h.UserService.GetRawUser(userID)
	if err != nil {
		respond.Err(c, apperr.NotFound("usuário não encontrado"))
		return
	}
	if user.TwoFAEnabled {
		respond.Err(c, apperr.Conflict("2FA já está ativado"))
		return
	}

	issuer := os.Getenv("APP_NAME")
	if issuer == "" {
		issuer = "5Vault"
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: user.Email,
	})
	if err != nil {
		respond.Err(c, apperr.Internal("falha ao gerar chave 2FA", err))
		return
	}

	// Salva o secret (ainda não habilitado — só ativa após confirmação)
	if err := h.UserService.Set2FASecret(userID, key.Secret()); err != nil {
		respond.Err(c, apperr.Internal("falha ao salvar secret", err))
		return
	}

	// Gera QR code como base64 PNG
	img, err := key.Image(200, 200)
	if err != nil {
		respond.Err(c, apperr.Internal("falha ao gerar QR code", err))
		return
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		respond.Err(c, apperr.Internal("falha ao encodar QR code", err))
		return
	}
	qr := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())

	respond.OK(c, gin.H{
		"secret":   key.Secret(),
		"qr_code":  qr,
		"otpauth":  key.URL(),
	})
}

// POST /user/2fa/verify — confirma o código e ativa o 2FA
func (h *UserHandler) Verify2FA(c *gin.Context) {
	userID := c.GetString("user_id")
	var body struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respond.Err(c, apperr.BadRequest("código obrigatório"))
		return
	}

	user, err := h.UserService.GetRawUser(userID)
	if err != nil || user.TwoFASecret == nil {
		respond.Err(c, apperr.BadRequest("configure o 2FA primeiro"))
		return
	}

	if !totp.Validate(body.Code, *user.TwoFASecret) {
		respond.Err(c, apperr.Unauthorized("código inválido"))
		return
	}

	if err := h.UserService.Enable2FA(userID); err != nil {
		respond.Err(c, apperr.Internal("falha ao ativar 2FA", err))
		return
	}

	respond.OK(c, gin.H{"message": "2FA ativado com sucesso"})
}

// POST /user/2fa/disable — desativa o 2FA
func (h *UserHandler) Disable2FA(c *gin.Context) {
	userID := c.GetString("user_id")
	var body struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respond.Err(c, apperr.BadRequest("código obrigatório"))
		return
	}

	user, err := h.UserService.GetRawUser(userID)
	if err != nil || user.TwoFASecret == nil || !user.TwoFAEnabled {
		respond.Err(c, apperr.BadRequest("2FA não está ativado"))
		return
	}

	if !totp.Validate(body.Code, *user.TwoFASecret) {
		respond.Err(c, apperr.Unauthorized("código inválido"))
		return
	}

	if err := h.UserService.Disable2FA(userID); err != nil {
		respond.Err(c, apperr.Internal("falha ao desativar 2FA", err))
		return
	}

	respond.OK(c, gin.H{"message": "2FA desativado"})
}
