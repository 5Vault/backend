package paymentHandlers

import (
	"backend/src/internal/domain/payment/services"
	"backend/src/pkg/apperr"
	"backend/src/pkg/respond"

	"github.com/gin-gonic/gin"
)

type CardHandler struct {
	Svc *paymentServices.CardService
}

func NewCardHandler(svc *paymentServices.CardService) *CardHandler {
	return &CardHandler{Svc: svc}
}

// POST /payment/cards — vincula um novo cartão via Stripe PaymentMethod ID
func (h *CardHandler) AttachCard(c *gin.Context) {
	var body struct {
		StripePMID string `json:"stripe_pm_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respond.Err(c, apperr.BadRequest("stripe_pm_id é obrigatório"))
		return
	}

	res, err := h.Svc.AttachCard(c.GetString("user_id"), body.StripePMID)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.Created(c, res)
}

// GET /payment/cards
func (h *CardHandler) ListCards(c *gin.Context) {
	res, err := h.Svc.ListCards(c.GetString("user_id"))
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, res)
}

// PATCH /payment/cards/:pmId/default
func (h *CardHandler) SetDefault(c *gin.Context) {
	if err := h.Svc.SetDefault(c.Param("pmId"), c.GetString("user_id")); err != nil {
		respond.Err(c, err)
		return
	}
	respond.NoContent(c)
}

// DELETE /payment/cards/:pmId
func (h *CardHandler) DeleteCard(c *gin.Context) {
	if err := h.Svc.DeleteCard(c.Param("pmId"), c.GetString("user_id")); err != nil {
		respond.Err(c, err)
		return
	}
	respond.NoContent(c)
}

// POST /lgpd/consent — registra o consentimento LGPD do usuário autenticado
func (h *CardHandler) RecordLGPDConsent(c *gin.Context) {
	if err := h.Svc.RecordLGPDConsent(c.GetString("user_id")); err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, gin.H{"message": "consentimento registrado"})
}
