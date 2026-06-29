package ticketHandlers

import (
	ticketSvc "backend/src/internal/domain/ticket/services"
	"backend/src/pkg/apperr"
	"backend/src/pkg/respond"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TicketHandler struct {
	Svc *ticketSvc.TicketService
}

func New(svc *ticketSvc.TicketService) *TicketHandler {
	return &TicketHandler{Svc: svc}
}

// POST /ticket
func (h *TicketHandler) Create(c *gin.Context) {
	var body struct {
		Subject  string `json:"subject" binding:"required,min=3,max=120"`
		Category string `json:"category"`
		Content  string `json:"content" binding:"required,min=10"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respond.Err(c, apperr.BadRequest(err.Error()))
		return
	}
	if body.Category == "" {
		body.Category = "outros"
	}
	t, err := h.Svc.Create(c.GetString("user_id"), body.Subject, body.Category, body.Content)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.Created(c, t)
}

// GET /ticket
func (h *TicketHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	tickets, total, err := h.Svc.ListForUser(c.GetString("user_id"), page, limit)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, gin.H{"tickets": tickets, "total": total})
}

// GET /ticket/:ticketId
func (h *TicketHandler) Get(c *gin.Context) {
	data, err := h.Svc.GetWithMessages(c.Param("ticketId"), c.GetString("user_id"), "user")
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, data)
}

// POST /ticket/:ticketId/reply
func (h *TicketHandler) Reply(c *gin.Context) {
	var body struct {
		Content string `json:"content" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respond.Err(c, apperr.BadRequest(err.Error()))
		return
	}
	msg, err := h.Svc.Reply(c.Param("ticketId"), c.GetString("user_id"), "user", body.Content)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.Created(c, msg)
}

// ── Admin ─────────────────────────────────────────────────────────────────────

// GET /admin/tickets
func (h *TicketHandler) AdminList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	tickets, total, err := h.Svc.ListAll(page, limit, status)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, gin.H{"tickets": tickets, "total": total})
}

// GET /admin/tickets/:ticketId
func (h *TicketHandler) AdminGet(c *gin.Context) {
	data, err := h.Svc.GetWithMessages(c.Param("ticketId"), c.GetString("user_id"), "admin")
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, data)
}

// POST /admin/tickets/:ticketId/reply
func (h *TicketHandler) AdminReply(c *gin.Context) {
	var body struct {
		Content string `json:"content" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respond.Err(c, apperr.BadRequest(err.Error()))
		return
	}
	msg, err := h.Svc.Reply(c.Param("ticketId"), c.GetString("user_id"), "admin", body.Content)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.Created(c, msg)
}

// PATCH /admin/tickets/:ticketId/close
func (h *TicketHandler) AdminClose(c *gin.Context) {
	if err := h.Svc.Close(c.Param("ticketId"), c.GetString("user_id")); err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, gin.H{"message": "ticket encerrado"})
}
