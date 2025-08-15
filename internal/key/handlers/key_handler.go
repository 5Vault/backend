package handlers

import (
	key "backend/internal/key/service"
	"backend/internal/models"
	_ "backend/internal/schemas"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create key"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Key created successfully"})
}

func (h *KeyHandler) ValidateKey(c *gin.Context) {
	var keyPayload *models.KeysPayload
	if err := c.ShouldBindJSON(&keyPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if !h.KeyService.FitKey(*keyPayload) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key format"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Key is valid"})
}
