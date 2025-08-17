package schemas

type File struct {
	ID         uint   `gorm:"primaryKey"`
	UserID     string `gorm:"index"`
	StorageID  string `gorm:"storage_id;index"`
	FileID     string `gorm:"file_id;index"`
	FileType   string `gorm:"file_type"`
	FileURL    string `gorm:"file_url"`
	UploadedAt string `gorm:"uploaded_at;default:CURRENT_TIMESTAMP"`
}
