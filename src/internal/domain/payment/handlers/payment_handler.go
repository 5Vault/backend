package paymentHandlers

import (
	"backend/src/external"
	"backend/src/internal/actionlog"
	"backend/src/internal/domain/payment/services"
	tierSvc "backend/src/internal/domain/user/tier/service"
	"backend/src/internal/logger"
	"backend/src/internal/schemas"
	paymentRepo "backend/src/internal/repository/payment"
	usrRepo "backend/src/internal/repository/user"
	"backend/src/pkg/apperr"
	"backend/src/pkg/respond"
	"backend/src/utils"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v82"
	"go.uber.org/zap"
)

type PaymentHandler struct {
	Svc         *paymentServices.PaymentService
	CardSvc     *paymentServices.CardService
	Tier        *tierSvc.TierService
	PaymentRepo *paymentRepo.PaymentRepository
	UserRepo    *usrRepo.UserRepository
	Email       *external.EmailClient
}

func NewPaymentHandler(svc *paymentServices.PaymentService, tier *tierSvc.TierService, pRepo *paymentRepo.PaymentRepository, uRepo *usrRepo.UserRepository, email *external.EmailClient) *PaymentHandler {
	return &PaymentHandler{Svc: svc, Tier: tier, PaymentRepo: pRepo, UserRepo: uRepo, Email: email}
}

func (h *PaymentHandler) WithCardService(cs *paymentServices.CardService) *PaymentHandler {
	h.CardSvc = cs
	return h
}

func (h *PaymentHandler) CreateIntent(c *gin.Context) {
	var body struct {
		TierID   string `json:"tier_id" binding:"required"`
		SaveCard bool   `json:"save_card"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respond.Err(c, apperr.BadRequest("tier_id is required"))
		return
	}

	cost := h.Tier.GetTierCostByID(body.TierID)
	if cost <= 0 {
		respond.Err(c, apperr.BadRequest("invalid or free tier"))
		return
	}

	secret, err := h.Svc.CreatePaymentIntent(c.GetString("user_id"), body.TierID, int64(cost*100), body.SaveCard)
	if err != nil {
		respond.Err(c, err)
		return
	}

	respond.OK(c, gin.H{
		"client_secret": secret,
		"amount_cents":  int64(cost * 100),
		"currency":      "brl",
	})
}

func (h *PaymentHandler) SaveCardFromIntent(c *gin.Context) {
	var body struct {
		PaymentIntentID string `json:"payment_intent_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respond.Err(c, apperr.BadRequest("payment_intent_id é obrigatório"))
		return
	}
	if h.CardSvc == nil {
		respond.Err(c, apperr.Internal("card service not configured", nil))
		return
	}

	pmID, err := h.Svc.GetPaymentMethodFromIntent(body.PaymentIntentID)
	if err != nil {
		respond.Err(c, apperr.BadRequest("payment intent não encontrado ou sem cartão"))
		return
	}

	res, err := h.CardSvc.AttachCard(c.GetString("user_id"), pmID)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.Created(c, res)
}

func (h *PaymentHandler) Webhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		respond.Err(c, apperr.BadRequest("cannot read body"))
		return
	}

	event, err := h.Svc.ValidateWebhook(payload, c.GetHeader("Stripe-Signature"))
	if err != nil {
		respond.Err(c, apperr.BadRequest("invalid webhook signature"))
		return
	}

	switch event.Type {
	case stripe.EventTypePaymentIntentSucceeded:
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			logger.Warn("failed to unmarshal payment intent", zap.Error(err))
			break
		}
		go h.handlePaymentSucceeded(pi)
	}

	respond.OK(c, gin.H{"received": true})
}

func (h *PaymentHandler) handlePaymentSucceeded(pi stripe.PaymentIntent) {
	userID := pi.Metadata["user_id"]
	tierID := pi.Metadata["tier_id"]
	if userID == "" || tierID == "" {
		logger.Warn("webhook missing metadata", zap.String("pi_id", pi.ID))
		return
	}

	if h.PaymentRepo.ExistsByStripeID(pi.ID) {
		return
	}

	now := time.Now()
	_ = h.PaymentRepo.Create(&schemas.Payment{
		PaymentID:   utils.GenerateULID(),
		UserID:      userID,
		StripeID:    pi.ID,
		TierID:      tierID,
		AmountCents: pi.Amount,
		Currency:    string(pi.Currency),
		Status:      schemas.PaymentStatusSucceeded,
		CreatedAt:   &now,
	})

	if err := h.UserRepo.SetUserTier(userID, tierID); err != nil {
		logger.Error("failed to upgrade tier", zap.String("user_id", userID), zap.String("tier", tierID), zap.Error(err))
		return
	}

	actionlog.Log(userID, "tier.upgrade", "payment", pi.ID, "", map[string]any{"tier": tierID, "amount": pi.Amount})
	logger.Info("tier upgraded", zap.String("user_id", userID), zap.String("tier", tierID))

	user, err := h.UserRepo.GetUserByID(userID)
	if err != nil {
		return
	}
	appURL := os.Getenv("APP_URL")
	_ = h.Email.RenderAndSend(user.Email, fmt.Sprintf("[FiveKeepr] Upgrade para %s confirmado!", tierID), "tier_upgrade", map[string]any{
		"TierName": tierID,
		"Amount":   fmt.Sprintf("%.2f", float64(pi.Amount)/100),
		"Date":     now.Format("02/01/2006"),
		"AppURL":   fmt.Sprintf("%s/dashboard", appURL),
	})
}

// GET /payment/history
func (h *PaymentHandler) ListPayments(c *gin.Context) {
	payments, _, err := h.PaymentRepo.ListByUserID(c.GetString("user_id"), 1, 50)
	if err != nil {
		respond.Err(c, apperr.Internal("erro ao listar pagamentos", err))
		return
	}
	respond.OK(c, gin.H{"payments": payments})
}
