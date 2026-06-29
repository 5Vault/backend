package paymentRepo

import (
	"backend/src/internal/schemas"

	"gorm.io/gorm"
)

type PaymentRepository struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{DB: db}
}

func (r *PaymentRepository) Create(p *schemas.Payment) error {
	return r.DB.Create(p).Error
}

func (r *PaymentRepository) ListByUserID(userID string, page, limit int) ([]schemas.Payment, int64, error) {
	var payments []schemas.Payment
	var total int64
	r.DB.Model(&schemas.Payment{}).Where("user_id = ?", userID).Count(&total)
	offset := (page - 1) * limit
	err := r.DB.Where("user_id = ?", userID).Order("created_at desc").Offset(offset).Limit(limit).Find(&payments).Error
	return payments, total, err
}

func (r *PaymentRepository) ExistsByStripeID(stripeID string) bool {
	var count int64
	r.DB.Model(&schemas.Payment{}).Where("stripe_id = ?", stripeID).Count(&count)
	return count > 0
}
