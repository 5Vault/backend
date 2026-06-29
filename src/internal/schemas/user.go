package schemas

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID            uint            `gorm:"primaryKey"`
	UserID        string          `gorm:"uniqueIndex;not null"`
	Username      string          `gorm:"uniqueIndex;not null"`
	Name          string
	Email         string          `gorm:"uniqueIndex;not null"`
	Password      string
	Phone         *string         `gorm:"uniqueIndex"`
	GoogleID      *string         `gorm:"uniqueIndex"`
	DiscordID     *string         `gorm:"uniqueIndex"`
	AvatarURL     *string
	AuthProvider  string          `gorm:"default:local"`
	Tier          string          `gorm:"default:free"`
	Role          string          `gorm:"default:user"`
	TierUpdatedAt        time.Time
	ExtraStorageEnabled  bool           `gorm:"default:false"`
	TwoFASecret          *string
	TwoFAEnabled         bool           `gorm:"default:false"`
	StripeCustomerID     *string        `gorm:"uniqueIndex"`
	LGPDConsentAt        *time.Time
	CreatedAt            *time.Time
	UpdatedAt            *time.Time
	DeletedAt            *gorm.DeletedAt `gorm:"index"`
}

func (User) TableName() string {
	return "user"
}
