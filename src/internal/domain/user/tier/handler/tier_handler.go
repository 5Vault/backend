package tHandler

import (
	"backend/src/internal/domain/user/tier/service"
	"backend/src/pkg/respond"

	"github.com/gin-gonic/gin"
)

type TierHandler struct {
	TierService *service.TierService
}

func NewTierHandler(service *service.TierService) *TierHandler {
	return &TierHandler{TierService: service}
}

func (h *TierHandler) GetTiers(c *gin.Context) {
	respond.OK(c, h.TierService.GetAllTiers())
}
