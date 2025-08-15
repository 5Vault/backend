package repository

import (
	"backend/internal/schemas"

	"gorm.io/gorm"
)

type KeyRepository struct {
	DB *gorm.DB
}

func NewKeyRepository(db *gorm.DB) *KeyRepository {
	return &KeyRepository{
		DB: db,
	}
}

func (repo *KeyRepository) New(key *schemas.Key) error {
	if err := repo.DB.Create(key).Error; err != nil {
		return err
	}
	return nil
}

func (repo *KeyRepository) GetByUserID(userID uint) (*schemas.Key, error) {
	var key *schemas.Key
	if err := repo.DB.First(&key, "user_id = ?", userID).Error; err != nil {
		return nil, err // Other error
	}
	return key, nil
}

func (repo *KeyRepository) GetByPublicKey(publicKey string) (*schemas.Key, error) {
	var key *schemas.Key
	if err := repo.DB.First(&key, "public_key = ?", publicKey).Error; err != nil {
		return nil, err // Other error
	}
	return key, nil
}

func (repo *KeyRepository) GetByPrivateKey(privateKey string) (*schemas.Key, error) {
	var key *schemas.Key
	if err := repo.DB.First(&key, "private_key = ?", privateKey).Error; err != nil {
		return nil, err // Other error
	}
	return key, nil
}
