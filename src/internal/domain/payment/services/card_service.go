package paymentServices

import (
	"backend/src/internal/logger"
	"backend/src/internal/models"
	pmRepo "backend/src/internal/repository/payment_method"
	"backend/src/internal/schemas"
	"backend/src/pkg/apperr"
	"backend/src/utils"
	"fmt"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/customer"
	stripePM "github.com/stripe/stripe-go/v82/paymentmethod"
	"go.uber.org/zap"
)

type CardService struct {
	Repo *pmRepo.PaymentMethodRepository
}

func NewCardService(repo *pmRepo.PaymentMethodRepository) *CardService {
	return &CardService{Repo: repo}
}

// AttachCard vincula um PaymentMethod do Stripe ao usuário e persiste apenas
// os metadados seguros (last4, brand, exp). CVV nunca passa pelo nosso servidor.
func (s *CardService) AttachCard(userID, stripePMID string) (*models.ResponsePaymentMethod, error) {
	// Busca ou cria o Customer no Stripe
	customerID, err := s.getOrCreateCustomer(userID)
	if err != nil {
		return nil, apperr.Internal("falha ao obter customer Stripe", err)
	}

	// Vincula o PaymentMethod ao Customer
	pm, err := stripePM.Attach(stripePMID, &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	})
	if err != nil {
		logger.Error("stripe attach failed", zap.String("user_id", userID), zap.Error(err))
		return nil, apperr.BadRequest("cartão inválido ou já vinculado")
	}

	if pm.Type != stripe.PaymentMethodTypeCard {
		return nil, apperr.BadRequest("apenas cartões são suportados")
	}

	// Verifica duplicata
	if _, err := s.Repo.GetByStripeID(pm.ID); err == nil {
		return nil, apperr.Conflict("este cartão já está cadastrado")
	}

	// Conta cartões existentes para definir se é o default
	existing, _ := s.Repo.ListByUserID(userID)
	isDefault := len(existing) == 0

	record := &schemas.PaymentMethod{
		PMID:      utils.GenerateULID(),
		UserID:    userID,
		StripeID:  pm.ID,
		CardLast4: pm.Card.Last4,
		CardBrand: string(pm.Card.Brand),
		ExpMonth:  int(pm.Card.ExpMonth),
		ExpYear:   int(pm.Card.ExpYear),
		IsDefault: isDefault,
	}

	if err := s.Repo.Create(record); err != nil {
		return nil, apperr.Internal("falha ao salvar cartão", err)
	}

	logger.Info("card attached",
		zap.String("user_id", userID),
		zap.String("last4", record.CardLast4),
		zap.String("brand", record.CardBrand),
	)
	return toResponse(record), nil
}

func (s *CardService) ListCards(userID string) ([]models.ResponsePaymentMethod, error) {
	pms, err := s.Repo.ListByUserID(userID)
	if err != nil {
		return nil, apperr.Internal("falha ao listar cartões", err)
	}
	result := make([]models.ResponsePaymentMethod, 0, len(pms))
	for i := range pms {
		result = append(result, *toResponse(&pms[i]))
	}
	return result, nil
}

func (s *CardService) SetDefault(pmID, userID string) error {
	if _, err := s.Repo.GetByID(pmID, userID); err != nil {
		return apperr.NotFound("cartão não encontrado")
	}
	return s.Repo.SetDefault(pmID, userID)
}

func (s *CardService) DeleteCard(pmID, userID string) error {
	pm, err := s.Repo.GetByID(pmID, userID)
	if err != nil {
		return apperr.NotFound("cartão não encontrado")
	}

	// Desvincula no Stripe
	if _, err := stripePM.Detach(pm.StripeID, nil); err != nil {
		logger.Warn("stripe detach failed", zap.String("stripe_id", pm.StripeID), zap.Error(err))
	}

	return s.Repo.Delete(pmID, userID)
}

func (s *CardService) RecordLGPDConsent(userID string) error {
	return s.Repo.RecordLGPDConsent(userID)
}

// ── interno ───────────────────────────────────────────────────────────────────

func (s *CardService) getOrCreateCustomer(userID string) (string, error) {
	existing, err := s.Repo.GetStripeCustomerID(userID)
	if err != nil {
		return "", err
	}
	if existing != "" {
		return existing, nil
	}

	cust, err := customer.New(&stripe.CustomerParams{
		Metadata: map[string]string{"user_id": userID},
	})
	if err != nil {
		return "", fmt.Errorf("stripe create customer: %w", err)
	}

	if err := s.Repo.UpdateStripeCustomerID(userID, cust.ID); err != nil {
		logger.Warn("failed to persist stripe customer id", zap.String("user_id", userID), zap.Error(err))
	}
	return cust.ID, nil
}

func toResponse(pm *schemas.PaymentMethod) *models.ResponsePaymentMethod {
	createdAt := ""
	if pm.CreatedAt != nil {
		createdAt = pm.CreatedAt.Format("2006-01-02T15:04:05Z")
	}
	return &models.ResponsePaymentMethod{
		PMID:      pm.PMID,
		CardLast4: pm.CardLast4,
		CardBrand: pm.CardBrand,
		ExpMonth:  pm.ExpMonth,
		ExpYear:   pm.ExpYear,
		IsDefault: pm.IsDefault,
		CreatedAt: createdAt,
	}
}
