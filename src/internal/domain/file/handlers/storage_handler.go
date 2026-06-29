package handlers

import (
	"backend/src/internal/domain/file/services"
	"backend/src/pkg/apperr"
	"backend/src/pkg/respond"
	"strconv"

	"github.com/gin-gonic/gin"
)

type StorageHandler struct {
	StorageService *services.StorageService
}

func NewStorageHandler(storageService *services.StorageService) *StorageHandler {
	return &StorageHandler{StorageService: storageService}
}

func (s *StorageHandler) GetFiles(c *gin.Context) {
	userID := c.GetString("user_id_key")
	if userID == "" {
		respond.Err(c, apperr.BadRequest("api key not provided"))
		return
	}

	itemsPerPage := 10
	if v, err := strconv.Atoi(c.Query("items_per_page")); err == nil && v > 0 {
		itemsPerPage = v
	}
	page := 1
	if v, err := strconv.Atoi(c.Query("page")); err == nil && v > 0 {
		page = v
	}

	files, err := s.StorageService.ListFiles(userID, itemsPerPage, page)
	if err != nil {
		respond.Err(c, apperr.Internal("failed to list files", err))
		return
	}
	if files == nil || len(*files) == 0 {
		respond.Err(c, apperr.NotFound("no files found"))
		return
	}
	respond.OK(c, files)
}

func (s *StorageHandler) GetFileStats(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		respond.Err(c, apperr.BadRequest("user_id not provided"))
		return
	}

	stats, err := s.StorageService.GetFileStats(userID)
	if err != nil {
		respond.Err(c, apperr.Internal("failed to get file stats", err))
		return
	}
	respond.OK(c, stats)
}

func (s *StorageHandler) GetFileByID(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		respond.Err(c, apperr.BadRequest("file_id not provided"))
		return
	}

	file, err := s.StorageService.GetFileByID(fileID)
	if err != nil {
		respond.Err(c, apperr.NotFound("file not found"))
		return
	}
	respond.OK(c, file)
}
