package handlers

import (
	"backend/src/internal/domain/file/services"
	"backend/src/internal/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type StorageHandler struct {
	StorageService *services.StorageService
}

func NewStorageHandler(storageService *services.StorageService) *StorageHandler {
	return &StorageHandler{
		StorageService: storageService,
	}
}

func (s *StorageHandler) UploadFile(c *gin.Context) {
	var UserID = c.GetString("user_id_key")

	var request *models.RequestFile
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	uploadResponse, err := s.StorageService.UploadFile(request, UserID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, uploadResponse)
}

func (s *StorageHandler) GetFiles(c *gin.Context) {
	var UserID = c.GetString("user_id_key")
	if UserID == "" {
		c.JSON(400, gin.H{"error": "api_key not provided"})
		return
	}
	itemsPerPageStr := c.Query("items_per_page")
	pageStr := c.Query("page")

	itemsPerPage, err := strconv.Atoi(itemsPerPageStr)
	if err != nil || itemsPerPage <= 0 {
		itemsPerPage = 10 // valor padrão
	}

	pageIndex := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			pageIndex = p
		}
	}

	file, err := s.StorageService.ListFiles(UserID, itemsPerPage, pageIndex)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if file == nil || len(*file) == 0 {
		c.JSON(404, gin.H{"message": "no files found"})
		return
	}
	c.JSON(200, file)
	return
}

func (s *StorageHandler) GetFileStats(c *gin.Context) {
	var UserID = c.GetString("user_id")
	if UserID == "" {
		c.JSON(400, gin.H{"error": "user_id not provided"})
		return
	}

	stats, err := s.StorageService.GetFileStats(UserID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, stats)
}
