package schemas

import (
	"time"

	"gorm.io/gorm"
)

type Storage struct {
	ID        uint            `gorm:"primaryKey"`
	UserID    uint            `gorm:"user_id;index"`
	Name      string          `gorm:"name;unique"`
	Size      int64           `gorm:"size"`
	CreatedAt *time.Time      `gorm:"autoCreateTime"`
	UpdatedAt *time.Time      `gorm:"autoUpdateTime"`
	DeletedAt *gorm.DeletedAt `gorm:"index"`
}
