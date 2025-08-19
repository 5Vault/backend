package schemas

import "time"

type File struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     string    `gorm:"index"`
	StorageID  string    `gorm:"storage_id;index"`
	FileID     string    `gorm:"file_id;index"`
	FileType   string    `gorm:"file_type"`
	FileURL    string    `gorm:"file_url"`
	UploadedAt time.Time `gorm:"uploaded_at"`
	FileSize   int64     `gorm:"file_size"`
}
