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

func (s *StorageHandler) GetFileByID(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(400, gin.H{"error": "file_id not provided"})
		return
	}

	// Verificar se o parâmetro details foi fornecido na query string
	detailsStr := c.Query("details")
	details := detailsStr == "true"

	file, err := s.StorageService.GetFileByID(fileID, details)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if file == nil {
		c.JSON(404, gin.H{"message": "file not found"})
		return
	}

	// Se details=false, retorna os dados binários com o Content-Type apropriado
	if !details {
		fileData, ok := file.(*models.FileData)
		if !ok {
			c.JSON(500, gin.H{"error": "erro ao processar dados do arquivo"})
			return
		}

		// Define o Content-Type correto para a imagem/arquivo
		c.Header("Content-Type", fileData.MimeType)
		c.Header("Content-Length", strconv.Itoa(len(fileData.Data)))

		// Retorna os dados binários diretamente
		c.Data(200, fileData.MimeType, fileData.Data)
		return
	}

	// Se details=true, retorna JSON com os metadados
	c.JSON(200, file)
}
