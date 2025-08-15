package handlers

import (
	key "backend/internal/key/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type KeyHandler struct {
	KeyService *key.Service
}

func NewKeyHandler(service *key.Service) *KeyHandler {
	return &KeyHandler{
		KeyService: service,
	}
}

func (h *KeyHandler) CreateKey(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	if err := h.KeyService.CreateKey(&userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Key created successfully"})
}

func (h *KeyHandler) ValidateKey(c *gin.Context) {
	apiKey := c.GetString("api_key")
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Api-Key header is required"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": apiKey})
}
