// Package actionlog provides async action logging throughout the application.
// Initialize once in routes with Init(db), then call Log() anywhere.
package actionlog

import (
	"backend/src/internal/logger"
	"backend/src/internal/schemas"
	"backend/src/utils"
	"encoding/json"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var db *gorm.DB

func Init(database *gorm.DB) {
	db = database
}

// Log records an action asynchronously so it never blocks the request.
func Log(userID, action, entityType, entityID, ip string, meta ...map[string]any) {
	if db == nil {
		return
	}
	metaStr := ""
	if len(meta) > 0 && meta[0] != nil {
		if b, err := json.Marshal(meta[0]); err == nil {
			metaStr = string(b)
		}
	}
	entry := &schemas.ActionLog{
		LogID:      utils.GenerateULID(),
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Meta:       metaStr,
		IP:         ip,
		CreatedAt:  func() *time.Time { t := time.Now(); return &t }(),
	}
	go func() {
		if err := db.Create(entry).Error; err != nil {
			logger.Warn("failed to write action log", zap.String("action", action), zap.Error(err))
		}
	}()
}
