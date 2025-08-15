package key

import (
	"backend/internal/key/repository"
	"backend/internal/models"
	"backend/internal/schemas"
	"backend/utils"
)

type Service struct {
	KeyRepo *repository.KeyRepository
}

func NewKeyService(repo *repository.KeyRepository) *Service {
	return &Service{
		KeyRepo: repo,
	}
}

func (s *Service) CreateKey(userId *string) error {

	keys, err := utils.GenerateKeyPair()
	if err != nil {
		return err
	}

	keySchema := &schemas.Key{
		PublicKey:  keys.PublicKey,
		PrivateKey: keys.PrivateKey,
		UserID:     *userId,
	}

	if err := s.KeyRepo.New(keySchema); err != nil {
		return err
	}
	return nil
}

func (s *Service) GetKeyByUserID(userID uint) (*schemas.Key, error) {
	key, err := s.KeyRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (s *Service) GetKeyByPublicKey(publicKey string) (*schemas.Key, error) {
	key, err := s.KeyRepo.GetByPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (s *Service) GetKeyByPrivateKey(privateKey string) (*schemas.Key, error) {
	key, err := s.KeyRepo.GetByPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (s *Service) FitKey(payload models.KeysPayload) bool {
	if payload.PublicKey == "" || payload.PrivateKey == "" {
		return false
	}

	key, err := s.GetKeyByPublicKey(payload.PublicKey)
	if err != nil {
		return false
	}

	if key == nil || key.PrivateKey != payload.PrivateKey {
		return false
	}

	return true
}
