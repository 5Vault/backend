package schemas

import "time"

// BackupBucket is a hidden R2 bucket per user used exclusively for backups.
type BackupBucket struct {
	ID        uint       `gorm:"primaryKey"`
	UserID    string     `gorm:"uniqueIndex;not null"`
	R2Name    string     `gorm:"uniqueIndex;not null"`
	CreatedAt *time.Time
}

func (BackupBucket) TableName() string { return "backup_bucket" }

// BackupSession represents one backup run triggered by the Lua script.
type BackupSession struct {
	SessionID  string     `gorm:"primaryKey"`
	KeyID      uint       `gorm:"index;not null"`
	UserID     string     `gorm:"index;not null"`
	Date       string     `gorm:"index;not null"` // YYYY-MM-DD (for daily quota)
	PathPrefix string     `gorm:"not null"`       // e.g. "2026-07-05/14-30-00"
	FileCount  int        `gorm:"default:0"`
	TotalSize  int64      `gorm:"default:0"`
	CreatedAt  *time.Time
}

func (BackupSession) TableName() string { return "backup_session" }
