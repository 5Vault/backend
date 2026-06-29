package notif

import (
	"backend/src/internal/schemas"
	"backend/src/utils"
	"time"

	"gorm.io/gorm"
)

var db *gorm.DB

func Init(database *gorm.DB) {
	db = database
}

// Create inserts a notification asynchronously.
func Create(userID, notifType, title, body, entityID string) {
	if db == nil {
		return
	}
	go func() {
		now := time.Now()
		_ = db.Create(&schemas.Notification{
			NotificationID: utils.GenerateULID(),
			UserID:         userID,
			Type:           notifType,
			Title:          title,
			Body:           body,
			EntityID:       entityID,
			CreatedAt:      &now,
		}).Error
	}()
}
