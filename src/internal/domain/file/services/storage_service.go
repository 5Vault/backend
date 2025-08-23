package services

import (
	"backend/src/internal/domain/file/repository"
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

var supa = external.NewSupaStorage()

func (s *StorageService) UploadFile(data *models.RequestFile, BucketID string) (*models.ResponseFile, error) {

	switch data.MimeType {
	case "image/png", "image/jpeg", "audio/mpeg", "video/mp4":
		// Tipos permitidos
	default:
		return nil, fmt.Errorf("tipo MIME não suportado: %s", data.MimeType)
	}
	res, err := supa.UploadFile(BucketID, data.Data, data.MimeType)
	if err != nil {
		return nil, err
	}

	newUrl := "https://" + os.Getenv("SUPABASE_ID") + ".supabase.co/storage/v1/object/public/" + res.Key

	file := &schemas.File{
		UserID:    BucketID,
		StorageID: BucketID,
		FileID:    res.Key,
		FileType:  data.MimeType,
		FileURL:   newUrl,
		FileSize:  int64(len(data.Data)),
	}

	fileID, err := s.Repo.CreateFile(file)
	if err != nil {
		return nil, fmt.Errorf("erro ao salvar arquivo no banco de dados: %v", err)
	}

	response := &models.ResponseFile{
		ID:         *fileID,
		UserID:     BucketID,
		StorageID:  BucketID,
		FileID:     res.Key,
		FileType:   data.MimeType,
		FileURL:    newUrl,
		UploadedAt: file.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
		FileSize:   int64(len(data.Data)),
	}

	fmt.Printf("DEBUG: Arquivo salvo com sucesso, ID: %d, URL: %s\n", *fileID, newUrl)
	return response, nil
}

func (s *StorageService) ListFiles(UserID string) (*[]models.ResponseFile, error) {
	files, err := s.Repo.GetFilesByUserID(UserID)
	if err != nil {
		return nil, err
	}
	var response []models.ResponseFile
	for _, file := range *files {
		response = append(response, models.ResponseFile{
			ID:         file.ID,
			FileID:     file.FileID,
			FileType:   file.FileType,
			FileURL:    file.FileURL,
			UserID:     file.UserID,
			StorageID:  file.StorageID,
			UploadedAt: file.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
			FileSize:   file.FileSize,
		})
	}
	return &response, nil
}

func (s *StorageService) GetFileStats(userID string) (*models.FileStats, error) {
	// Obter estatísticas básicas do repository (dados brutos)
	totalFiles, usedSize, err := s.Repo.GetFileStats(userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter estatísticas dos arquivos: %v", err)
	}

	// Tamanho total do storage fixo de 250 GB (lógica de negócio no service)
	const totalStorageBytes int64 = 250 * 1024 * 1024 * 1024 // 250 GB em bytes

	// Calcular espaço livre (lógica de negócio no service)
	freeSpace := totalStorageBytes - usedSize
	if freeSpace < 0 {
		freeSpace = 0 // Garantir que não seja negativo
	}

	// Obter arquivos recentes (últimos 5) - apenas dados brutos
	recentFilesData, err := s.Repo.GetRecentFilesByUserID(userID, 5)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter arquivos recentes: %v", err)
	}

	// Converter schemas.File para models.ResponseFile (responsabilidade do service)
	var recentFiles []models.ResponseFile
	if recentFilesData != nil {
		for _, file := range *recentFilesData {
			recentFiles = append(recentFiles, models.ResponseFile{
				ID:         file.ID,
				FileID:     file.FileID,
				FileType:   file.FileType,
				FileURL:    file.FileURL,
				UserID:     file.UserID,
				StorageID:  file.StorageID,
				UploadedAt: file.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
				FileSize:   file.FileSize,
			})
		}
	}

	// Montar o modelo de resposta final (responsabilidade do service)
	return &models.FileStats{
		TotalFiles:  totalFiles,
		UsedSize:    usedSize,          // Tamanho real dos arquivos existentes
		TotalSize:   totalStorageBytes, // 250 GB fixo
		FreeSpace:   freeSpace,         // Espaço disponível
		RecentFiles: recentFiles,
	}, nil
}

//func (s *StorageService) DeleteFile() error {
//	// Implement your services logic here
//	return nil
//}
//
//func (s *StorageService) GetFile() ([]byte, error) {
//	// Implement your services logic here
//	return nil, nil
//}
