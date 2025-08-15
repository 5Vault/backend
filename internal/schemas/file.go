package schemas

type File struct {
	ID        uint   `gorm:"primaryKey"`
	StorageID uint   `gorm:"storage_id;index"`
	FileID    uint   `gorm:"file_id;index"`
	FileType  string `gorm:"file_type"`
	FileURL   string `gorm:"file_url"`
}
