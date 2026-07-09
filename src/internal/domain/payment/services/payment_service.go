package paymentServices

import (
	"backend/src/internal/logger"
	"fmt"
	"os"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/webhook"
	"go.uber.org/zap"
)

func init() {
	key := os.Getenv("STRIPE_SECRET_KEY")
	if key == "" {
		fmt.Println("STRIPE_SECRET_KEY not set — payment features disabled")
		return
	}
	stripe.Key = key
}

type PaymentService struct{}

func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

func (s *PaymentService) CreatePaymentIntent(userID, tierID string, amountCents int64, saveCard bool) (string, error) {
	params := &stripe.PaymentIntentParams{
		Amount:             stripe.Int64(amountCents),
		Currency:           stripe.String(string(stripe.CurrencyBRL)),
		PaymentMethodTypes: []*string{stripe.String("card")},
		Metadata: map[string]string{
			"user_id":   userID,
			"tier_id":   tierID,
			"save_card": fmt.Sprintf("%v", saveCard),
		},
	}
	if saveCard {
		params.SetupFutureUsage = stripe.String("off_session")
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		logger.Error("failed to create payment intent", zap.String("user_id", userID), zap.Error(err))
		return "", fmt.Errorf("failed to create payment intent: %w", err)
	}

	logger.Info("payment intent created",
		zap.String("user_id", userID),
		zap.String("tier_id", tierID),
		zap.String("pi_id", pi.ID),
	)
	return pi.ClientSecret, nil
}

// GetPaymentMethodFromIntent retrieves the payment method ID used in a
// completed payment intent so it can be saved for future use.
func (s *PaymentService) GetPaymentMethodFromIntent(paymentIntentID string) (string, error) {
	pi, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		return "", fmt.Errorf("get payment intent: %w", err)
	}
	if pi.PaymentMethod == nil || pi.PaymentMethod.ID == "" {
		return "", fmt.Errorf("no payment method on intent %s", paymentIntentID)
	}
	return pi.PaymentMethod.ID, nil
}

func (s *PaymentService) ValidateWebhook(payload []byte, sigHeader string) (*stripe.Event, error) {
	secret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	event, err := webhook.ConstructEvent(payload, sigHeader, secret)
	if err != nil {
		return nil, fmt.Errorf("invalid webhook signature: %w", err)
	}
	return &event, nil
}
