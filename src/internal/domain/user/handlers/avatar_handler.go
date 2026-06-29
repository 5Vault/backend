package handlers

import (
	"backend/src/pkg/apperr"
	"backend/src/pkg/respond"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

const maxAvatarSize = 2 << 20 // 2 MB

var allowedAvatarTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID := c.GetString("user_id")

	fh, err := c.FormFile("avatar")
	if err != nil {
		respond.Err(c, apperr.BadRequest("arquivo de avatar não encontrado"))
		return
	}

	if fh.Size > maxAvatarSize {
		respond.Err(c, apperr.BadRequest("avatar muito grande (máx 2 MB)"))
		return
	}

	ext, ok := detectAvatarType(fh)
	if !ok {
		respond.Err(c, apperr.BadRequest("tipo de arquivo não permitido (use JPG, PNG, WEBP ou GIF)"))
		return
	}

	dir := "./uploads/avatars"
	if err := os.MkdirAll(dir, 0755); err != nil {
		respond.Err(c, apperr.Internal("falha ao criar diretório", err))
		return
	}

	// Remove avatar anterior do mesmo usuário (qualquer extensão)
	for _, e := range allowedAvatarTypes {
		_ = os.Remove(filepath.Join(dir, userID+e))
	}

	filename := userID + ext
	dst := filepath.Join(dir, filename)
	if err := c.SaveUploadedFile(fh, dst); err != nil {
		respond.Err(c, apperr.Internal("falha ao salvar avatar", err))
		return
	}

	baseURL := os.Getenv("SERVER_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	avatarURL := fmt.Sprintf("%s/uploads/avatars/%s", strings.TrimRight(baseURL, "/"), filename)

	if err := h.UserService.UpdateAvatar(userID, avatarURL); err != nil {
		respond.Err(c, apperr.Internal("falha ao atualizar avatar", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"avatar_url": avatarURL})
}

func detectAvatarType(fh *multipart.FileHeader) (string, bool) {
	f, err := fh.Open()
	if err != nil {
		return "", false
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	ct := http.DetectContentType(buf[:n])
	ext, ok := allowedAvatarTypes[ct]
	return ext, ok
}
