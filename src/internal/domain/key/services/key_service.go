package key

import (
	"backend/src/internal/repository/key"
	"backend/src/internal/schemas"
	"backend/src/utils"
	"fmt"
)

type Service struct {
	KeyRepo *key.KeyRepository
}

func NewKeyService(repo *key.KeyRepository) *Service {
	return &Service{
		KeyRepo: repo,
	}
}

func (s *Service) CreateKey(userId *string) error {

	if userId == nil {
		return fmt.Errorf("userId is nil")
	}

	key, err := utils.GenerateAPIKey()
	if err != nil {
		return err
	}

	keySchema := &schemas.Key{
		Key:    key,
		UserID: *userId,
	}

	if err := s.KeyRepo.New(keySchema); err != nil {
		return err
	}
	return nil
}

func (s *Service) GetKeyByUserID(userID string) (*schemas.Key, error) {
	key, err := s.KeyRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (s *Service) ValidateKey(key string) (*schemas.Key, error) {
	result, err := s.KeyRepo.GetKey(key)
	if err != nil {
		return nil, fmt.Errorf("error retrieving key: %w", err)
	}
	if result == nil {
		return nil, fmt.Errorf("key not found")
	}
	return result, nil
}
