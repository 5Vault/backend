package schemas

import "time"

type PasswordResetToken struct {
	TokenID   string     `gorm:"primaryKey"`
	UserID    string     `gorm:"index;not null"`
	Token     string     `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time  `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt *time.Time
}

func (PasswordResetToken) TableName() string { return "password_reset_tokens" }
