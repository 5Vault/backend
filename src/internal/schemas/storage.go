package schemas

import (
	"time"

	"gorm.io/gorm"
)

// Directory é um diretório (prefixo) dentro de um Bucket.
// Arquivos ficam em {DirID}/{filename} no bucket R2.
type Directory struct {
	DirID     string          `gorm:"primaryKey"`
	BucketID  string          `gorm:"index:idx_dir_bucket_user,priority:1;not null"`
	UserID    string          `gorm:"index:idx_dir_bucket_user,priority:2;not null"`
	Name      string          `gorm:"not null"`
	CreatedAt *time.Time
	DeletedAt *gorm.DeletedAt `gorm:"index"`
}

func (Directory) TableName() string { return "directory" }
