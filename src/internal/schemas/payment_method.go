package schemas

import "time"

// PaymentMethod armazena APENAS metadados do cartão vindos do Stripe.
// NUNCA armazena PAN completo, CVV ou dados brutos — apenas o pm_xxx do Stripe
// e os campos que o próprio Stripe expõe de forma segura (last4, brand, exp).
type PaymentMethod struct {
	PMID       string     `gorm:"primaryKey"`
	UserID     string     `gorm:"index;not null"`
	StripeID   string     `gorm:"uniqueIndex;not null"` // pm_xxx retornado pelo Stripe
	CardLast4  string     `gorm:"not null"`
	CardBrand  string     `gorm:"not null"` // visa, mastercard, elo, etc.
	ExpMonth   int        `gorm:"not null"`
	ExpYear    int        `gorm:"not null"`
	IsDefault  bool       `gorm:"default:false"`
	CreatedAt  *time.Time
}

func (PaymentMethod) TableName() string { return "payment_method" }
