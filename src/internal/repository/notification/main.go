package notificationRepo

import (
	"backend/src/internal/schemas"
	"time"

	"gorm.io/gorm"
)

type NotificationRepository struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{DB: db}
}

func (r *NotificationRepository) Create(n *schemas.Notification) error {
	return r.DB.Create(n).Error
}

// ListByUserID returns notifications newest-first, paginated.
func (r *NotificationRepository) ListByUserID(userID string, page, limit int) ([]schemas.Notification, int64, error) {
	var items []schemas.Notification
	var total int64
	r.DB.Model(&schemas.Notification{}).Where("user_id = ?", userID).Count(&total)
	offset := (page - 1) * limit
	err := r.DB.Where("user_id = ?", userID).
		Order("created_at desc").
		Offset(offset).Limit(limit).
		Find(&items).Error
	return items, total, err
}

// UnreadCountByType returns a map of type → unread count for a user.
func (r *NotificationRepository) UnreadCountByType(userID string) (map[string]int64, error) {
	type row struct {
		Type  string
		Count int64
	}
	var rows []row
	err := r.DB.Model(&schemas.Notification{}).
		Select("type, count(*) as count").
		Where("user_id = ? AND read_at IS NULL", userID).
		Group("type").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]int64)
	for _, r := range rows {
		result[r.Type] = r.Count
	}
	return result, nil
}

func (r *NotificationRepository) MarkRead(notificationID, userID string) error {
	now := time.Now()
	return r.DB.Model(&schemas.Notification{}).
		Where("notification_id = ? AND user_id = ?", notificationID, userID).
		Update("read_at", &now).Error
}

func (r *NotificationRepository) MarkAllRead(userID string) error {
	now := time.Now()
	return r.DB.Model(&schemas.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", &now).Error
}

// MarkReadByEntity marks all unread notifications of a given type+entityID as read.
func (r *NotificationRepository) MarkReadByEntity(userID, notifType, entityID string) error {
	now := time.Now()
	return r.DB.Model(&schemas.Notification{}).
		Where("user_id = ? AND type = ? AND entity_id = ? AND read_at IS NULL", userID, notifType, entityID).
		Update("read_at", &now).Error
}
