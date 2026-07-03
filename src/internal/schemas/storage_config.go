package schemas

import (
	"time"

	"gorm.io/gorm"
)

type BucketStatus string

const (
	BucketStatusPending BucketStatus = "pending"
	BucketStatusActive  BucketStatus = "active"
	BucketStatusError   BucketStatus = "error"
)

// Bucket representa um bucket R2 real na Cloudflare, de propriedade de um usuário.
// O nome no R2 é sempre fk-{BucketID}.
type Bucket struct {
	BucketID            string          `gorm:"primaryKey"`
	UserID              string          `gorm:"index:idx_bucket_user_status,priority:1;not null"`
	Name                string          `gorm:"not null"`
	R2Name              string          `gorm:"uniqueIndex;not null"`
	Status              BucketStatus    `gorm:"index:idx_bucket_user_status,priority:2;default:pending"`
	FileCount           int64           `gorm:"default:0"`
	BytesUsed           int64           `gorm:"default:0"`
	CustomDomain        string          `gorm:"default:''"`
	PublicDomain        string          `gorm:"default:''"`
	PublicAccessEnabled bool            `gorm:"default:false"`
	CreatedAt           *time.Time
	DeletedAt           *gorm.DeletedAt `gorm:"index"`
}

func (Bucket) TableName() string { return "bucket" }
