package notificationHandlers

import (
	notifRepo "backend/src/internal/repository/notification"
	"backend/src/pkg/apperr"
	"backend/src/pkg/respond"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	Repo *notifRepo.NotificationRepository
}

func New(repo *notifRepo.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{Repo: repo}
}

// GET /notifications
func (h *NotificationHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	items, total, err := h.Repo.ListByUserID(c.GetString("user_id"), page, limit)
	if err != nil {
		respond.Err(c, apperr.Internal("erro ao listar notificações", err))
		return
	}
	respond.OK(c, gin.H{"notifications": items, "total": total})
}

// GET /notifications/unread-count
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	counts, err := h.Repo.UnreadCountByType(c.GetString("user_id"))
	if err != nil {
		respond.Err(c, apperr.Internal("erro ao contar notificações", err))
		return
	}
	respond.OK(c, gin.H{"counts": counts})
}

// PATCH /notifications/:id/read
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	if err := h.Repo.MarkRead(c.Param("id"), c.GetString("user_id")); err != nil {
		respond.Err(c, apperr.Internal("erro ao marcar lida", err))
		return
	}
	respond.OK(c, gin.H{"ok": true})
}

// POST /notifications/read-all
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	if err := h.Repo.MarkAllRead(c.GetString("user_id")); err != nil {
		respond.Err(c, apperr.Internal("erro ao marcar todas como lidas", err))
		return
	}
	respond.OK(c, gin.H{"ok": true})
}

// POST /notifications/read-entity  { type, entity_id }
func (h *NotificationHandler) MarkReadByEntity(c *gin.Context) {
	var body struct {
		Type     string `json:"type" binding:"required"`
		EntityID string `json:"entity_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respond.Err(c, apperr.BadRequest(err.Error()))
		return
	}
	if err := h.Repo.MarkReadByEntity(c.GetString("user_id"), body.Type, body.EntityID); err != nil {
		respond.Err(c, apperr.Internal("erro", err))
		return
	}
	respond.OK(c, gin.H{"ok": true})
}
