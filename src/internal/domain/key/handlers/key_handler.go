package handlers

import (
	"backend/src/internal/domain/key/services"
	"backend/src/internal/schemas"
	"backend/src/pkg/apperr"
	"backend/src/pkg/respond"
	"strconv"

	"github.com/gin-gonic/gin"
)

type KeyHandler struct {
	KeyService *key.Service
}

func NewKeyHandler(service *key.Service) *KeyHandler {
	return &KeyHandler{KeyService: service}
}

func (h *KeyHandler) CreateKey(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		respond.Err(c, apperr.Unauthorized("usuário não autenticado"))
		return
	}

	var body struct {
		Label       string `json:"label"`
		Permission  string `json:"permission"`
		AllBuckets  bool   `json:"all_buckets"`
		BucketPerms []struct {
			BucketID   string `json:"bucket_id"`
			Permission string `json:"permission"`
		} `json:"bucket_perms"`
	}
	_ = c.ShouldBindJSON(&body)

	bps := make([]key.BucketPerm, 0, len(body.BucketPerms))
	for _, bp := range body.BucketPerms {
		bps = append(bps, key.BucketPerm{
			BucketID:   bp.BucketID,
			Permission: schemas.KeyPermission(bp.Permission),
		})
	}

	if err := h.KeyService.CreateKey(&userID, body.Label, schemas.KeyPermission(body.Permission), body.AllBuckets, bps); err != nil {
		respond.Err(c, err)
		return
	}

	respond.Created(c, gin.H{"message": "chave criada com sucesso"})
}

func (h *KeyHandler) ListKeys(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		respond.Err(c, apperr.Unauthorized("usuário não autenticado"))
		return
	}

	keys, err := h.KeyService.ListKeys(userID)
	if err != nil {
		respond.Err(c, err)
		return
	}

	type bucketPermRow struct {
		BucketID   string `json:"bucket_id"`
		Permission string `json:"permission"`
	}

	type keyRow struct {
		ID          uint            `json:"id"`
		Label       string          `json:"label"`
		Key         string          `json:"key"`
		Permission  string          `json:"permission"`
		AllBuckets  bool            `json:"all_buckets"`
		BucketPerms []bucketPermRow `json:"bucket_perms"`
		CreatedAt   string          `json:"created_at"`
	}

	rows := make([]keyRow, 0, len(keys))
	for _, k := range keys {
		ca := ""
		if k.CreatedAt != nil {
			ca = k.CreatedAt.Format("2006-01-02 15:04:05")
		}
		bps := make([]bucketPermRow, 0, len(k.BucketPerms))
		for _, bp := range k.BucketPerms {
			bps = append(bps, bucketPermRow{BucketID: bp.BucketID, Permission: string(bp.Permission)})
		}
		rows = append(rows, keyRow{
			ID:          k.ID,
			Label:       k.Label,
			Key:         k.Key,
			Permission:  string(k.Permission),
			AllBuckets:  k.AllBuckets,
			BucketPerms: bps,
			CreatedAt:   ca,
		})
	}

	respond.OK(c, rows)
}

func (h *KeyHandler) DeleteKey(c *gin.Context) {
	userID := c.GetString("user_id")
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		respond.Err(c, apperr.BadRequest("id inválido"))
		return
	}

	if err := h.KeyService.DeleteKey(uint(id), userID); err != nil {
		respond.Err(c, err)
		return
	}

	respond.NoContent(c)
}

func (h *KeyHandler) ValidateKey(c *gin.Context) {
	respond.OK(c, gin.H{"valid_key": true})
}
