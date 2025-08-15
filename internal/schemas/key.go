package schemas

import "time"

type Key struct {
	ID         uint       `gorm:"primaryKey"`
	UserID     string     `gorm:"user_id;index"`
	PublicKey  string     `gorm:"public_key;unique"`
	PrivateKey string     `gorm:"private_key;unique"`
	CreatedAt  *time.Time `gorm:"autoCreateTime"`
}
