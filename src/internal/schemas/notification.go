package schemas

import "time"

type Notification struct {
	NotificationID string     `gorm:"primaryKey"`
	UserID         string     `gorm:"index;not null"`
	Type           string     `gorm:"not null"` // "ticket_reply", "system", "tier_upgrade"
	Title          string     `gorm:"not null"`
	Body           string     `gorm:"default:''"`
	EntityID       string     `gorm:"default:''"` // ticket_id, etc.
	ReadAt         *time.Time
	CreatedAt      *time.Time
}

func (Notification) TableName() string { return "notifications" }
