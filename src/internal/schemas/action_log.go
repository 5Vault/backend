package schemas

import "time"

type ActionLog struct {
	LogID      string     `gorm:"primaryKey"`
	UserID     string     `gorm:"index;not null"`
	Action     string     `gorm:"not null"` // ex: "login", "file.upload", "tier.upgrade"
	EntityType string     `gorm:"default:''"`
	EntityID   string     `gorm:"default:''"`
	Meta       string     `gorm:"type:text;default:''"` // JSON
	IP         string     `gorm:"default:''"`
	CreatedAt  *time.Time
}

func (ActionLog) TableName() string { return "action_logs" }
