package tHandler

import (
	"backend/src/internal/domain/user/tier/service"

	"github.com/gin-gonic/gin"
)

// TierHandler provides HTTP handlers related to tiers.
type TierHandler struct {
	TierService *service.TierService
}

func NewTierHandler(service *service.TierService) *TierHandler {
	return &TierHandler{
		TierService: service,
	}
}

// GetTiers writes the available tiers as JSON to the response.
func (h *TierHandler) GetTiers(c *gin.Context) {
	c.JSON(200, h.TierService.GetAllTiers())
}

// GetUserString For backwards compatibility, if some code expects a method that returns a string,
// keep a simple helper (not a Gin handler).
func (h *TierHandler) GetUserString() string {
	return "tier"
}
