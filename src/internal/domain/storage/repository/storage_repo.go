package repository

import (
	"backend/src/internal/models"
	"backend/src/internal/schemas"

	"gorm.io/gorm"
)

type StorageRepository struct {
	DB *gorm.DB
}

func NewStorageRepository(db *gorm.DB) *StorageRepository {
	return &StorageRepository{
		DB: db,
	}
}

func (repo *StorageRepository) CreateFile(file *schemas.File) error {
	if err := repo.DB.Create(file).Error; err != nil {
		return err
	}
	return nil
}

func (repo *StorageRepository) GetFileByID(fileID uint) (*models.File, error) {
	var file *models.File
	if err := repo.DB.First(&file, "file_id = ?", fileID).Error; err != nil {
		return nil, err // Other error
	}
	return file, nil
}

func (repo *StorageRepository) GetFilesByUserID(userID string) (*[]schemas.File, error) {
	var files *[]schemas.File
	if err := repo.DB.Where("user_id = ?", userID).Find(&files).Error; err != nil {
		return nil, err // Other error
	}
	return files, nil
}

func (repo *StorageRepository) GetFilesByKey(key string) ([]models.File, error) {
	var files []models.File
	if err := repo.DB.Where("key = ?", key).Find(&files).Error; err != nil {
		return nil, err // Other error
	}
	return files, nil
}

func (repo *StorageRepository) GetFilesByStorageID(storageID uint) ([]models.File, error) {
	var files []models.File
	if err := repo.DB.Where("storage_id = ?", storageID).Find(&files).Error; err != nil {
		return nil, err // Other error
	}
	return files, nil
}
