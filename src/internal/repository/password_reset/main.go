package passwordResetRepo

import (
	"backend/src/internal/schemas"
	"time"

	"gorm.io/gorm"
)

type PasswordResetRepository struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *PasswordResetRepository {
	return &PasswordResetRepository{DB: db}
}

func (r *PasswordResetRepository) Create(t *schemas.PasswordResetToken) error {
	return r.DB.Create(t).Error
}

func (r *PasswordResetRepository) GetValid(token string) (*schemas.PasswordResetToken, error) {
	var t schemas.PasswordResetToken
	err := r.DB.Where("token = ? AND used_at IS NULL AND expires_at > ?", token, time.Now()).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *PasswordResetRepository) MarkUsed(tokenID string) error {
	now := time.Now()
	return r.DB.Model(&schemas.PasswordResetToken{}).Where("token_id = ?", tokenID).Update("used_at", now).Error
}

func (r *PasswordResetRepository) InvalidateByUserID(userID string) error {
	now := time.Now()
	return r.DB.Model(&schemas.PasswordResetToken{}).
		Where("user_id = ? AND used_at IS NULL", userID).
		Update("used_at", now).Error
}
