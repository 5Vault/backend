package services

import (
	"backend/src/internal/logger"
	"backend/src/internal/models"
	"backend/src/internal/repository/file"
	"fmt"

	"go.uber.org/zap"
)

type StorageService struct {
	Repo *file.StorageRepository
}

func NewStorageService(repo *file.StorageRepository) *StorageService {
	return &StorageService{Repo: repo}
}

func (s *StorageService) ListFiles(userID string, itemsPerPage, page int) (*[]models.ResponseFile, error) {
	files, err := s.Repo.GetFilesByUserID(userID, itemsPerPage, page)
	if err != nil {
		logger.Error("failed to list files", zap.String("user_id", userID), zap.Error(err))
		return nil, err
	}

	var response []models.ResponseFile
	for _, f := range *files {
		fileURL := f.FileURL
		if fileURL == "" {
			fileURL = "/api/v1/file/" + f.FileID
		}
		response = append(response, models.ResponseFile{
			ID:         f.ID,
			FileID:     f.FileID,
			FileType:   f.FileType,
			FileURL:    fileURL,
			UserID:     f.UserID,
			StorageID:  f.StorageID,
			UploadedAt: f.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
			FileSize:   f.FileSize,
		})
	}
	return &response, nil
}

func (s *StorageService) GetFileStats(userID string) (*models.FileStats, error) {
	totalFiles, usedSize, err := s.Repo.GetFileStats(userID)
	if err != nil {
		logger.Error("failed to get file stats", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("erro ao obter estatísticas: %w", err)
	}

	const totalStorageBytes int64 = 250 * 1024 * 1024 * 1024
	freeSpace := totalStorageBytes - usedSize
	if freeSpace < 0 {
		freeSpace = 0
	}

	recentFilesData, err := s.Repo.GetRecentFilesByUserID(userID, 5)
	if err != nil {
		logger.Error("failed to get recent files", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("erro ao obter arquivos recentes: %w", err)
	}

	var recentFiles []models.ResponseFile
	if recentFilesData != nil {
		for _, f := range *recentFilesData {
			fileURL := f.FileURL
			if fileURL == "" {
				fileURL = "/api/v1/file/" + f.FileID
			}
			recentFiles = append(recentFiles, models.ResponseFile{
				ID:         f.ID,
				FileID:     f.FileID,
				FileType:   f.FileType,
				FileURL:    fileURL,
				UserID:     f.UserID,
				StorageID:  f.StorageID,
				UploadedAt: f.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
				FileSize:   f.FileSize,
			})
		}
	}

	weeklyUsage, err := s.Repo.GetWeeklyFileUsage(userID)
	if err != nil {
		logger.Error("failed to get weekly usage", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("erro ao obter uso semanal: %w", err)
	}

	return &models.FileStats{
		TotalFiles:  totalFiles,
		UsedSize:    usedSize,
		TotalSize:   totalStorageBytes,
		FreeSpace:   freeSpace,
		RecentFiles: recentFiles,
		WeeklyUsage: weeklyUsage,
	}, nil
}

func (s *StorageService) GetFileByID(fileID string) (*models.ResponseFile, error) {
	f, err := s.Repo.GetFileByID(fileID)
	if err != nil {
		logger.Warn("file not found", zap.String("file_id", fileID), zap.Error(err))
		return nil, err
	}

	return &models.ResponseFile{
		ID:         f.ID,
		FileID:     f.FileID,
		FileType:   f.FileType,
		FileURL:    f.FileURL,
		UserID:     f.UserID,
		StorageID:  f.StorageID,
		UploadedAt: f.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
		FileSize:   f.FileSize,
	}, nil
}
