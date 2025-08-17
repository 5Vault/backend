package services

import (
	"backend/src/internal/domain/storage/repository"
	"backend/src/internal/models"
	"backend/src/internal/schemas"
	"fmt"
	"os"
)
import "backend/src/external"

type StorageService struct {
	Repo *repository.StorageRepository
}

func NewStorageService(repo *repository.StorageRepository) *StorageService {
	return &StorageService{
		Repo: repo,
	}
}

var supa external.SupaStorage

func (s *StorageService) UploadFile(data *models.RequestFile, UserID string) (*string, error) {
	switch data.MimeType {
	case "image/png", "image/jpeg", "audio/mpeg", "video/mp4":
	// Tipos permitidos
	default:
		return nil, fmt.Errorf("tipo MIME não suportado: %s", data.MimeType)
	}

	createBucket, _ := supa.GetBucket(UserID)
	if createBucket == nil {
		_, err := supa.CreateBucket(UserID)
		if err != nil {
			_ = fmt.Errorf("erro ao criar bucket: %v", err)
		}
	}

	res, err := supa.UploadFile(UserID, data.Data, data.MimeType)
	if err != nil {
		return nil, err
	}
	newUrl := "https://" + os.Getenv("SUPABASE_ID") + ".supabase.co/storage/v1/object/public/" + res.Key

	file := &schemas.File{
		UserID:    UserID,
		StorageID: UserID,
		FileID:    res.Key,
		FileType:  data.MimeType,
		FileURL:   newUrl,
	}

	if err := s.Repo.CreateFile(file); err != nil {
		return nil, fmt.Errorf("erro ao salvar arquivo no banco de dados: %v", err)
	}
	return &newUrl, err
}

func (s *StorageService) ListFiles() ([]string, error) {
	// Implement your services logic here
	return nil, nil
}

func (s *StorageService) DeleteFile() error {
	// Implement your services logic here
	return nil
}

func (s *StorageService) GetFile() ([]byte, error) {
	// Implement your services logic here
	return nil, nil
}
