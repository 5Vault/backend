package key

import (
	"backend/src/internal/repository/key"
	"backend/src/internal/schemas"
	"backend/src/pkg/apperr"
	"backend/src/utils"
)

const maxKeysPerUser = 10

type Service struct {
	KeyRepo *key.KeyRepository
}

func NewKeyService(repo *key.KeyRepository) *Service {
	return &Service{KeyRepo: repo}
}

type BucketPerm struct {
	BucketID   string
	Permission schemas.KeyPermission
}

func (s *Service) CreateKey(userID *string, label string, permission schemas.KeyPermission, allBuckets bool, bucketPerms []BucketPerm) error {
	if userID == nil {
		return apperr.BadRequest("user id é obrigatório")
	}

	count, _ := s.KeyRepo.CountByUserID(*userID)
	if count >= maxKeysPerUser {
		return apperr.BadRequest("limite de 10 chaves por usuário atingido")
	}

	if permission == "" {
		permission = schemas.KeyPermissionReadWrite
	}
	if permission != schemas.KeyPermissionRead && permission != schemas.KeyPermissionReadWrite {
		return apperr.BadRequest("permissão inválida: use 'read' ou 'readwrite'")
	}

	k, err := utils.GenerateAPIKey()
	if err != nil {
		return apperr.Internal("falha ao gerar chave", err)
	}

	key := &schemas.Key{
		Key:        k,
		UserID:     *userID,
		Label:      label,
		Permission: permission,
		AllBuckets: allBuckets || len(bucketPerms) == 0,
	}
	if err := s.KeyRepo.New(key); err != nil {
		return apperr.Internal("falha ao criar chave", err)
	}

	for _, bp := range bucketPerms {
		perm := bp.Permission
		if perm == "" {
			perm = permission
		}
		if err := s.KeyRepo.AddBucketPerm(&schemas.KeyBucketPermission{
			KeyID:      key.ID,
			BucketID:   bp.BucketID,
			Permission: perm,
		}); err != nil {
			return apperr.Internal("falha ao salvar permissão de bucket", err)
		}
	}
	return nil
}

func (s *Service) ListKeys(userID string) ([]schemas.Key, error) {
	keys, err := s.KeyRepo.ListByUserID(userID)
	if err != nil {
		return nil, apperr.Internal("falha ao listar chaves", err)
	}
	return keys, nil
}

func (s *Service) DeleteKey(id uint, userID string) error {
	if err := s.KeyRepo.DeleteByID(id, userID); err != nil {
		return apperr.NotFound("chave não encontrada")
	}
	return nil
}

func (s *Service) GetKeyByUserID(userID string) (*schemas.Key, error) {
	k, err := s.KeyRepo.GetByUserID(userID)
	if err != nil {
		return nil, apperr.NotFound("chave não encontrada")
	}
	return k, nil
}

func (s *Service) ValidateKey(k string) (*schemas.Key, error) {
	result, err := s.KeyRepo.GetKey(k)
	if err != nil || result == nil {
		return nil, apperr.Unauthorized("chave de API inválida")
	}
	return result, nil
}
