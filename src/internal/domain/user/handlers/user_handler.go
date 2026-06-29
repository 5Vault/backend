package handlers

import (
	usrSvc "backend/src/internal/domain/user/services"
	"backend/src/internal/models"
	"backend/src/pkg/apperr"
	"backend/src/pkg/respond"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	UserService *usrSvc.UserService
}

func NewUserHandler(userService *usrSvc.UserService) *UserHandler {
	return &UserHandler{UserService: userService}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.RequestUser
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Err(c, apperr.BadRequest(err.Error()))
		return
	}
	newID, err := h.UserService.CreateUser(&req)
	if err != nil {
		respond.Err(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "user created", "user_id": newID})
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	userID := c.Param("id")
	user, err := h.UserService.GetUserByID(userID, false)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, user)
}

func (h *UserHandler) GetOwnUser(c *gin.Context) {
	userID := c.GetString("user_id")
	user, err := h.UserService.GetUserByID(userID, true)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, user)
}

func (h *UserHandler) SetExtraStorage(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Err(c, apperr.BadRequest(err.Error()))
		return
	}
	if err := h.UserService.SetExtraStorage(c.GetString("user_id"), req.Enabled); err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, gin.H{"extra_storage_enabled": req.Enabled})
}
