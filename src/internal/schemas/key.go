package schemas

import "time"

type Key struct {
	ID        uint       `gorm:"primaryKey"`
	UserID    string     `gorm:"user_id;index"`
	Key       string     `gorm:"key"`
	CreatedAt *time.Time `gorm:"autoCreateTime"`
}

func (Key) TableName() string {
	return "key"
}
