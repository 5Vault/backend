package storageRepo

import (
	"backend/src/internal/schemas"

	"gorm.io/gorm"
)

type DirectoryRepository struct {
	DB *gorm.DB
}

func NewStorageRepository(db *gorm.DB) *DirectoryRepository {
	return &DirectoryRepository{DB: db}
}

func (r *DirectoryRepository) Create(d *schemas.Directory) error {
	return r.DB.Create(d).Error
}

func (r *DirectoryRepository) ListByBucket(bucketID, userID string) ([]schemas.Directory, error) {
	var dirs []schemas.Directory
	err := r.DB.Where("bucket_id = ? AND user_id = ?", bucketID, userID).Order("created_at desc").Find(&dirs).Error
	return dirs, err
}

func (r *DirectoryRepository) GetByID(dirID, bucketID, userID string) (*schemas.Directory, error) {
	var d schemas.Directory
	if err := r.DB.Where("dir_id = ? AND bucket_id = ? AND user_id = ?", dirID, bucketID, userID).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DirectoryRepository) GetByName(name, bucketID, userID string) (*schemas.Directory, error) {
	var d schemas.Directory
	if err := r.DB.Where("name = ? AND bucket_id = ? AND user_id = ?", name, bucketID, userID).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DirectoryRepository) Delete(dirID, bucketID, userID string) error {
	return r.DB.Where("dir_id = ? AND bucket_id = ? AND user_id = ?", dirID, bucketID, userID).Delete(&schemas.Directory{}).Error
}
