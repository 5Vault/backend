package handlers

import (
	"backend/src/internal/domain/storage/services"
	"backend/src/internal/models"

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
	c.JSON(200, gin.H{"data": uploadResponse})
	return
}

func (s *StorageHandler) GetFiles(c *gin.Context) {
	var UserID = c.GetString("api_key")
	if UserID == "" {
		c.JSON(400, gin.H{"error": "api_key not provided"})
		return
	}

	file, err := s.StorageService.ListFiles(UserID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, file)
	return
}
