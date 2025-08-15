package handlers

import (
	"backend/src/internal/domain/user/services"
	"backend/src/internal/models"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	UserService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		UserService: userService,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var request models.RequestUser
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	newId, err := h.UserService.CreateUser(&request)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"message": "User created successfully",
		"user_id": &newId,
	})
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(400, gin.H{"error": "User ID is required"})
		return
	}

	user, err := h.UserService.GetUserByID(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if user == nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	c.JSON(200, user)
}

func (h *UserHandler) GetOwnUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(400, gin.H{"error": "User ID not found in context"})
		return
	}

	user, err := h.UserService.GetUserByID(userID.(string))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if user == nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	c.JSON(200, user)
}
