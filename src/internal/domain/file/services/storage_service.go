package services

import (
	"backend/src/internal/models"
	"backend/src/internal/repository/file"
	"backend/src/internal/schemas"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)
import "backend/src/external"

type StorageService struct {
	Repo *file.StorageRepository
}

func NewStorageService(repo *file.StorageRepository) *StorageService {
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
	fileName := uuid.New().String()
	res, err := supa.UploadFile(BucketID, data.Data, data.MimeType, fileName)
	if err != nil {
		return nil, err
	}

	newUrl := "https://" + os.Getenv("SUPABASE_ID") + ".supabase.co/storage/v1/object/public/" + res.Key

	fileRes := &schemas.File{
		UserID:     BucketID,
		StorageID:  BucketID,
		FileID:     fileName,
		FileType:   data.MimeType,
		FileURL:    newUrl,
		FileSize:   int64(len(data.Data)),
		UploadedAt: time.Now(),
	}

	fileID, err := s.Repo.CreateFile(fileRes)
	if err != nil {
		return nil, fmt.Errorf("erro ao salvar arquivo no banco de dados: %v", err)
	}

	response := &models.ResponseFile{
		ID:         *fileID,
		UserID:     BucketID,
		StorageID:  BucketID,
		FileID:     fileName,
		FileType:   data.MimeType,
		FileURL:    "/api/v1/file/" + fileName,
		UploadedAt: fileRes.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
		FileSize:   int64(len(data.Data)),
	}

	return response, nil
}

func (s *StorageService) ListFiles(UserID string, ItemsPerPage, Page int) (*[]models.ResponseFile, error) {
	files, err := s.Repo.GetFilesByUserID(UserID, ItemsPerPage, Page)
	if err != nil {
		return nil, err
	}
	var response []models.ResponseFile
	for _, file := range *files {
		response = append(response, models.ResponseFile{
			ID:         file.ID,
			FileID:     file.FileID,
			FileType:   file.FileType,
			FileURL:    "/api/v1/file/" + file.FileID,
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
				FileURL:    "/api/v1/file/" + file.FileID,
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

func (s *StorageService) GetFileByID(fileID string, details bool) (interface{}, error) {
	fileConsult, err := s.Repo.GetFileByID(fileID)
	if err != nil {
		return nil, err
	}

	// Se details=false, busca e retorna os dados binários do arquivo do storage
	if !details {
		// Busca o arquivo real do Supabase Storage
		fileData, err := supa.DownloadFile(fileConsult.FileURL)
		if err != nil {
			return nil, fmt.Errorf("erro ao buscar arquivo do storage: %v", err)
		}

		fileRes := &models.FileData{
			Data:     fileData,
			MimeType: fileConsult.FileType,
		}
		return fileRes, nil
	}

	// Se details=true, retorna o payload completo com todos os detalhes
	fileRes := &models.ResponseFile{
		ID:         fileConsult.ID,
		FileID:     fileConsult.FileID,
		FileType:   fileConsult.FileType,
		FileURL:    "/api/v1/file/" + fileConsult.FileID,
		UserID:     fileConsult.UserID,
		StorageID:  fileConsult.StorageID,
		UploadedAt: fileConsult.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
		FileSize:   fileConsult.FileSize,
	}
	return fileRes, nil
}
