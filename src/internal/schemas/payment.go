package schemas

import "time"

type PaymentStatus string

const (
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

type Payment struct {
	PaymentID   string        `gorm:"primaryKey"`
	UserID      string        `gorm:"index;not null"`
	StripeID    string        `gorm:"uniqueIndex;not null"` // pi_xxx
	TierID      string        `gorm:"not null"`
	AmountCents int64         `gorm:"not null"`
	Currency    string        `gorm:"default:'brl'"`
	Status      PaymentStatus `gorm:"default:'succeeded'"`
	CreatedAt   *time.Time
}

func (Payment) TableName() string { return "payments" }
