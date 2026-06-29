package pmRepo

import (
	"backend/src/internal/schemas"
	"gorm.io/gorm"
)

type PaymentMethodRepository struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *PaymentMethodRepository {
	return &PaymentMethodRepository{DB: db}
}

func (r *PaymentMethodRepository) Create(pm *schemas.PaymentMethod) error {
	return r.DB.Create(pm).Error
}

func (r *PaymentMethodRepository) ListByUserID(userID string) ([]schemas.PaymentMethod, error) {
	var pms []schemas.PaymentMethod
	err := r.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&pms).Error
	return pms, err
}

func (r *PaymentMethodRepository) GetByID(pmID, userID string) (*schemas.PaymentMethod, error) {
	var pm schemas.PaymentMethod
	if err := r.DB.Where("pm_id = ? AND user_id = ?", pmID, userID).First(&pm).Error; err != nil {
		return nil, err
	}
	return &pm, nil
}

func (r *PaymentMethodRepository) GetByStripeID(stripeID string) (*schemas.PaymentMethod, error) {
	var pm schemas.PaymentMethod
	if err := r.DB.Where("stripe_id = ?", stripeID).First(&pm).Error; err != nil {
		return nil, err
	}
	return &pm, nil
}

func (r *PaymentMethodRepository) SetDefault(pmID, userID string) error {
	tx := r.DB.Begin()
	if err := tx.Model(&schemas.PaymentMethod{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&schemas.PaymentMethod{}).Where("pm_id = ? AND user_id = ?", pmID, userID).Update("is_default", true).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (r *PaymentMethodRepository) Delete(pmID, userID string) error {
	return r.DB.Where("pm_id = ? AND user_id = ?", pmID, userID).Delete(&schemas.PaymentMethod{}).Error
}

func (r *PaymentMethodRepository) UpdateStripeCustomerID(userID, customerID string) error {
	return r.DB.Model(&schemas.User{}).Where("user_id = ?", userID).Update("stripe_customer_id", customerID).Error
}

func (r *PaymentMethodRepository) GetStripeCustomerID(userID string) (string, error) {
	var user schemas.User
	if err := r.DB.Select("stripe_customer_id").Where("user_id = ?", userID).First(&user).Error; err != nil {
		return "", err
	}
	if user.StripeCustomerID == nil {
		return "", nil
	}
	return *user.StripeCustomerID, nil
}

func (r *PaymentMethodRepository) RecordLGPDConsent(userID string) error {
	return r.DB.Model(&schemas.User{}).Where("user_id = ?", userID).
		UpdateColumn("lgpd_consent_at", gorm.Expr("NOW()")).Error
}
