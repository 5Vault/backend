package schemas

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    string `gorm:"unique;not null"`
	Username  string `gorm:"unique"`
	Name      string `gorm:"name"`
	Email     string `gorm:"unique"`
	Password  string `gorm:"password"`
	Phone     string `gorm:"phone"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt *gorm.DeletedAt `gorm:"index"`
}
