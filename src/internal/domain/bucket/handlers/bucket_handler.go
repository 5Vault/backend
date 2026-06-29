package bucketHandlers

import (
	bucketSvc "backend/src/internal/domain/bucket/services"
	"backend/src/internal/models"
	"backend/src/pkg/apperr"
	"backend/src/pkg/respond"
	"bytes"
	"io"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var allowedMIMEPrefixes = []string{"image/", "audio/", "video/"}

var allowedExtensions = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
	".bmp": true, ".ico": true, ".tiff": true, ".tif": true, ".svg": true,
	".mp3": true, ".wav": true, ".ogg": true, ".flac": true, ".aac": true,
	".m4a": true, ".opus": true, ".wma": true,
	".mp4": true, ".webm": true, ".mkv": true, ".avi": true, ".mov": true,
	".wmv": true, ".flv": true, ".m4v": true, ".3gp": true,
}

func isMediaFile(filename, contentType string) bool {
	for _, p := range allowedMIMEPrefixes {
		if strings.HasPrefix(contentType, p) {
			return true
		}
	}
	if dot := strings.LastIndex(filename, "."); dot >= 0 {
		return allowedExtensions[strings.ToLower(filename[dot:])]
	}
	return false
}

type BucketHandler struct {
	Svc *bucketSvc.BucketService
}

func NewBucketHandler(svc *bucketSvc.BucketService) *BucketHandler {
	return &BucketHandler{Svc: svc}
}

// GET /bucket/stats
func (h *BucketHandler) GetStats(c *gin.Context) {
	res, err := h.Svc.GetStats(c.GetString("user_id"))
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, res)
}

// PATCH /bucket/:bucketId/domain
func (h *BucketHandler) SetDomain(c *gin.Context) {
	var req models.RequestSetDomain
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Err(c, apperr.BadRequest(err.Error()))
		return
	}
	if err := h.Svc.SetCustomDomain(c.Request.Context(), c.Param("bucketId"), c.GetString("user_id"), req.Domain); err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, gin.H{"message": "domínio atualizado com sucesso"})
}

// POST /bucket/:bucketId/public-access
func (h *BucketHandler) EnablePublicAccess(c *gin.Context) {
	domain, err := h.Svc.EnablePublicAccess(c.Request.Context(), c.Param("bucketId"), c.GetString("user_id"))
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, gin.H{"public_domain": domain})
}

// POST /bucket/
func (h *BucketHandler) CreateBucket(c *gin.Context) {
	var req models.RequestCreateBucket
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Err(c, apperr.BadRequest(err.Error()))
		return
	}
	res, err := h.Svc.CreateBucket(c.GetString("user_id"), req.Name)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.Created(c, res)
}

// GET /bucket/
func (h *BucketHandler) ListBuckets(c *gin.Context) {
	res, err := h.Svc.ListBuckets(c.GetString("user_id"))
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, res)
}

// DELETE /bucket/:bucketId
func (h *BucketHandler) DeleteBucket(c *gin.Context) {
	if err := h.Svc.DeleteBucket(c.Param("bucketId"), c.GetString("user_id")); err != nil {
		respond.Err(c, err)
		return
	}
	respond.NoContent(c)
}

// POST /bucket/:bucketId/dir
func (h *BucketHandler) CreateDirectory(c *gin.Context) {
	var req models.RequestCreateDirectory
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Err(c, apperr.BadRequest(err.Error()))
		return
	}
	res, err := h.Svc.CreateDirectory(c.Param("bucketId"), c.GetString("user_id"), req.Name)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.Created(c, res)
}

// GET /bucket/:bucketId/dir
func (h *BucketHandler) ListDirectories(c *gin.Context) {
	res, err := h.Svc.ListDirectories(c.Param("bucketId"), c.GetString("user_id"))
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, res)
}

// DELETE /bucket/:bucketId/dir/:dirId
func (h *BucketHandler) DeleteDirectory(c *gin.Context) {
	if err := h.Svc.DeleteDirectory(c.Param("dirId"), c.Param("bucketId"), c.GetString("user_id")); err != nil {
		respond.Err(c, err)
		return
	}
	respond.NoContent(c)
}

// POST /bucket/:bucketId/dir/:dirId/files
func (h *BucketHandler) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		respond.Err(c, apperr.BadRequest("file é obrigatório"))
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if !isMediaFile(header.Filename, contentType) {
		respond.Err(c, apperr.BadRequest("apenas arquivos de mídia são permitidos (imagem, áudio ou vídeo)"))
		return
	}

	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, file); err != nil {
		respond.Err(c, apperr.Internal("failed to read file", err))
		return
	}

	res, err := h.Svc.UploadFile(c.Request.Context(), c.Param("bucketId"), c.Param("dirId"), c.GetString("user_id"), header.Filename, buf.Bytes(), contentType)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.Created(c, res)
}

// POST /bucket/:bucketId/upload
func (h *BucketHandler) UploadFilePublic(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		respond.Err(c, apperr.BadRequest("file é obrigatório"))
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if !isMediaFile(header.Filename, contentType) {
		respond.Err(c, apperr.BadRequest("apenas arquivos de mídia são permitidos (imagem, áudio ou vídeo)"))
		return
	}

	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, file); err != nil {
		respond.Err(c, apperr.Internal("failed to read file", err))
		return
	}

	forceCreate := c.Query("force_create") == "true"
	userID := c.GetString("user_id_key")
	if userID == "" {
		userID = c.GetString("user_id")
	}

	res, err := h.Svc.UploadFilePublic(c.Request.Context(), c.Param("bucketId"), userID, header.Filename, buf.Bytes(), contentType, forceCreate)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.Created(c, res)
}

// GET /bucket/:bucketId/files
func (h *BucketHandler) ListFilesPublic(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	dirName := c.DefaultQuery("dir", "")
	
	userID := c.GetString("user_id_key")
	if userID == "" {
		userID = c.GetString("user_id")
	}

	res, err := h.Svc.ListFilesPublic(c.Request.Context(), c.Param("bucketId"), userID, dirName, page, limit)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, res)
}

// DELETE /bucket/:bucketId/files
func (h *BucketHandler) DeleteFilePublic(c *gin.Context) {
	filename := c.DefaultQuery("filename", "")
	if filename == "" {
		filename = c.Param("filename")
	}
	if filename == "" {
		respond.Err(c, apperr.BadRequest("filename é obrigatório"))
		return
	}

	userID := c.GetString("user_id_key")
	if userID == "" {
		userID = c.GetString("user_id")
	}

	if err := h.Svc.DeleteFilePublic(c.Request.Context(), c.Param("bucketId"), userID, filename); err != nil {
		respond.Err(c, err)
		return
	}
	respond.NoContent(c)
}

// GET /bucket/:bucketId/dir/:dirId/files?page=1&limit=20
func (h *BucketHandler) ListFiles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	res, err := h.Svc.ListFiles(c.Request.Context(), c.Param("bucketId"), c.Param("dirId"), c.GetString("user_id"), page, limit)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, res)
}

// DELETE /bucket/:bucketId/dir/:dirId/files/:filename
func (h *BucketHandler) DeleteFile(c *gin.Context) {
	if err := h.Svc.DeleteFile(c.Request.Context(), c.Param("bucketId"), c.Param("dirId"), c.GetString("user_id"), c.Param("filename")); err != nil {
		respond.Err(c, err)
		return
	}
	respond.NoContent(c)
}
