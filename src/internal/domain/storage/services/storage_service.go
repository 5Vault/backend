package services

import (
	"backend/src/internal/models"
	"fmt"
)
import "backend/src/external"

type StorageService struct{}

func NewStorageService() *StorageService {
	return &StorageService{}
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
	return &res.Key, err
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
