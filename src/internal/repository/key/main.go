package key

import (
	"backend/src/internal/schemas"

	"gorm.io/gorm"
)

type KeyRepository struct {
	DB *gorm.DB
}

func NewKeyRepository(db *gorm.DB) *KeyRepository {
	return &KeyRepository{DB: db}
}

func (repo *KeyRepository) New(key *schemas.Key) error {
	return repo.DB.Create(key).Error
}

func (repo *KeyRepository) GetByUserID(userID string) (*schemas.Key, error) {
	var key schemas.Key
	if err := repo.DB.Where("user_id = ?", userID).First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (repo *KeyRepository) ListByUserID(userID string) ([]schemas.Key, error) {
	var keys []schemas.Key
	if err := repo.DB.Preload("BucketPerms").Where("user_id = ?", userID).Order("created_at desc").Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

func (repo *KeyRepository) AddBucketPerm(perm *schemas.KeyBucketPermission) error {
	return repo.DB.Create(perm).Error
}

func (repo *KeyRepository) DeleteBucketPerms(keyID uint) error {
	return repo.DB.Where("key_id = ?", keyID).Delete(&schemas.KeyBucketPermission{}).Error
}

func (repo *KeyRepository) GetKey(key string) (*schemas.Key, error) {
	var result schemas.Key
	if err := repo.DB.Where("key = ?", key).First(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (repo *KeyRepository) DeleteByID(id uint, userID string) error {
	return repo.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&schemas.Key{}).Error
}

func (repo *KeyRepository) CountByUserID(userID string) (int64, error) {
	var count int64
	repo.DB.Model(&schemas.Key{}).Where("user_id = ?", userID).Count(&count)
	return count, nil
}
