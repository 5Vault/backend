package oauthHandlers

import (
	oauthSvc "backend/src/internal/domain/oauth/services"
	"backend/src/pkg/apperr"
	"backend/src/pkg/respond"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type OAuthHandler struct {
	OAuthService        *oauthSvc.OAuthService
	DiscordOAuthService *oauthSvc.DiscordOAuthService
}

func NewOAuthHandler(service *oauthSvc.OAuthService) *OAuthHandler {
	return &OAuthHandler{OAuthService: service}
}

func (h *OAuthHandler) WithDiscord(svc *oauthSvc.DiscordOAuthService) *OAuthHandler {
	h.DiscordOAuthService = svc
	return h
}

func (h *OAuthHandler) GoogleLogin(c *gin.Context) {
	url := h.OAuthService.GetGoogleAuthURL("fivevault-oauth-state")
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *OAuthHandler) GoogleCallback(c *gin.Context) {
	if c.Query("state") != "fivevault-oauth-state" {
		respond.Err(c, apperr.BadRequest("invalid oauth state"))
		return
	}
	code := c.Query("code")
	if code == "" {
		respond.Err(c, apperr.BadRequest("authorization code not provided"))
		return
	}

	token, err := h.OAuthService.HandleGoogleCallback(c.Request.Context(), code)
	if err != nil {
		respond.Err(c, err)
		return
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/account?mode=login&token=%s", frontendURL, token))
}

func (h *OAuthHandler) DiscordLogin(c *gin.Context) {
	if h.DiscordOAuthService == nil {
		respond.Err(c, apperr.Internal("discord oauth not configured"))
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, h.DiscordOAuthService.GetAuthURL("fivevault-discord-state"))
}

func (h *OAuthHandler) DiscordCallback(c *gin.Context) {
	frontendURL := os.Getenv("FRONTEND_URL")

	if h.DiscordOAuthService == nil {
		respond.Err(c, apperr.Internal("discord oauth not configured"))
		return
	}

	// Usuário cancelou ou houve erro no lado do Discord
	if errParam := c.Query("error"); errParam != "" {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/account?mode=login", frontendURL))
		return
	}

	if c.Query("state") != "fivevault-discord-state" {
		respond.Err(c, apperr.BadRequest("invalid oauth state"))
		return
	}
	code := c.Query("code")
	if code == "" {
		respond.Err(c, apperr.BadRequest("authorization code not provided"))
		return
	}

	token, err := h.DiscordOAuthService.HandleCallback(c.Request.Context(), code)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/account?mode=login", frontendURL))
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/account?mode=login&token=%s", frontendURL, token))
}
