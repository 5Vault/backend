package backupHandlers

import (
	backupSvc "backend/src/internal/domain/backup/services"
	"backend/src/pkg/apperr"
	"backend/src/pkg/respond"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BackupHandler struct {
	BackupService *backupSvc.BackupService
}

func New(svc *backupSvc.BackupService) *BackupHandler {
	return &BackupHandler{BackupService: svc}
}

// POST /api/v1/backup/file
// Auth: Api-Key middleware (sets user_id_key, key_id)
// Body: multipart with fields: file, path, session_id (optional)
func (h *BackupHandler) UploadFile(c *gin.Context) {
	userID := c.GetString("user_id_key")
	keyID := uint(c.GetUint("key_id"))

	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		respond.Err(c, apperr.BadRequest("invalid multipart form"))
		return
	}

	filePath := c.PostForm("path")
	if filePath == "" {
		respond.Err(c, apperr.BadRequest("campo 'path' é obrigatório"))
		return
	}

	sessionID := c.PostForm("session_id")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		respond.Err(c, apperr.BadRequest("campo 'file' é obrigatório"))
		return
	}

	if fileHeader.Size > 50<<20 { // 50 MB
		respond.Err(c, apperr.BadRequest("arquivo muito grande (máx 50 MB)"))
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		respond.Err(c, apperr.Internal("failed to open file"))
		return
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		respond.Err(c, apperr.Internal("failed to read file"))
		return
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		ext := filepath.Ext(filePath)
		if ext != "" {
			if mt := mime.TypeByExtension(ext); mt != "" {
				contentType = mt
			}
		}
		if contentType == "" {
			contentType = http.DetectContentType(data)
		}
	}

	result, err := h.BackupService.UploadFile(c.Request.Context(), keyID, userID, sessionID, filePath, data, contentType)
	if err != nil {
		respond.Err(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GET /api/v1/backup/quota
// Auth: Api-Key middleware
func (h *BackupHandler) GetQuota(c *gin.Context) {
	userID := c.GetString("user_id_key")
	keyID := uint(c.GetUint("key_id"))

	result, err := h.BackupService.GetQuota(keyID, userID)
	if err != nil {
		respond.Err(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// GET /api/v1/backup/sessions
// Auth: JWT middleware
// Query: date (YYYY-MM-DD), page, limit
func (h *BackupHandler) ListSessions(c *gin.Context) {
	userID := c.GetString("user_id")

	date := c.Query("date")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	sessions, total, err := h.BackupService.ListSessions(userID, date, page, limit)
	if err != nil {
		respond.Err(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}
