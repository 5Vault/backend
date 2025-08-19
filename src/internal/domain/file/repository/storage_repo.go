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

func (repo *StorageRepository) CreateFile(file *schemas.File) (*uint, error) {
	if err := repo.DB.Create(file).Error; err != nil {
		return nil, err
	}
	return &file.ID, nil
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

func (repo *StorageRepository) GetFilesByKey(key string) (*[]schemas.File, error) {
	var files *[]schemas.File
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

func (repo *StorageRepository) GetFileStats(userID string) (int64, int64, error) {
	var result struct {
		TotalFiles int64
		TotalSize  int64
	}

	if err := repo.DB.Model(&schemas.File{}).
		Select("COUNT(*) as total_files, COALESCE(SUM(file_size), 0) as total_size").
		Where("user_id = ?", userID).
		Scan(&result).Error; err != nil {
		return 0, 0, err
	}

	return result.TotalFiles, result.TotalSize, nil
}

func (repo *StorageRepository) GetRecentFilesByUserID(userID string, limit int) (*[]schemas.File, error) {
	var files *[]schemas.File
	if err := repo.DB.Where("user_id = ?", userID).
		Order("uploaded_at DESC").
		Limit(limit).
		Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}
