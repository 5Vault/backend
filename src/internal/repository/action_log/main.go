package actionLogRepo

import (
	"backend/src/internal/schemas"

	"gorm.io/gorm"
)

type ActionLogRepository struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *ActionLogRepository {
	return &ActionLogRepository{DB: db}
}

func (r *ActionLogRepository) Create(l *schemas.ActionLog) error {
	return r.DB.Create(l).Error
}

func (r *ActionLogRepository) ListByUserID(userID string, page, limit int) ([]schemas.ActionLog, int64, error) {
	var logs []schemas.ActionLog
	var total int64
	r.DB.Model(&schemas.ActionLog{}).Where("user_id = ?", userID).Count(&total)
	offset := (page - 1) * limit
	err := r.DB.Where("user_id = ?", userID).Order("created_at desc").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}

func (r *ActionLogRepository) HasLoginFromIP(userID, ip string) bool {
	var count int64
	r.DB.Model(&schemas.ActionLog{}).
		Where("user_id = ? AND action = 'login' AND ip = ?", userID, ip).
		Count(&count)
	return count > 0
}
