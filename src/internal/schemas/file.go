package schemas

type File struct {
	ID         uint   `gorm:"primaryKey"`
	Key        string `gorm:"key;index"`
	UserID     string `gorm:"index"`
	StorageID  uint   `gorm:"storage_id;index"`
	FileID     uint   `gorm:"file_id;index"`
	FileType   string `gorm:"file_type"`
	FileURL    string `gorm:"file_url"`
	UploadedAt string `gorm:"uploaded_at;default:CURRENT_TIMESTAMP"`
}
