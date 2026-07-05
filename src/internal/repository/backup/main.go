package backupRepo

import (
	"backend/src/internal/schemas"
	"time"

	"gorm.io/gorm"
)

type BackupRepository struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *BackupRepository {
	return &BackupRepository{DB: db}
}

// ── BackupBucket ──────────────────────────────────────────────────────────────

func (r *BackupRepository) GetBucketByUserID(userID string) (*schemas.BackupBucket, error) {
	var b schemas.BackupBucket
	if err := r.DB.Where("user_id = ?", userID).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BackupRepository) CreateBucket(b *schemas.BackupBucket) error {
	return r.DB.Create(b).Error
}

// ── BackupSession ─────────────────────────────────────────────────────────────

func (r *BackupRepository) CreateSession(s *schemas.BackupSession) error {
	return r.DB.Create(s).Error
}

func (r *BackupRepository) CountTodayByKey(keyID uint, date string) (int64, error) {
	var count int64
	err := r.DB.Model(&schemas.BackupSession{}).
		Where("key_id = ? AND date = ?", keyID, date).
		Count(&count).Error
	return count, err
}

func (r *BackupRepository) IncrStats(sessionID string, files int, size int64) {
	r.DB.Model(&schemas.BackupSession{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]any{
			"file_count": gorm.Expr("file_count + ?", files),
			"total_size": gorm.Expr("total_size + ?", size),
		})
}

func (r *BackupRepository) ListByUserID(userID string, date string, page, limit int) ([]schemas.BackupSession, int64, error) {
	var sessions []schemas.BackupSession
	var total int64

	q := r.DB.Model(&schemas.BackupSession{}).Where("user_id = ?", userID)
	if date != "" {
		q = q.Where("date = ?", date)
	}

	q.Count(&total)
	err := q.Order("created_at desc").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&sessions).Error
	return sessions, total, err
}

func (r *BackupRepository) GetSession(sessionID, userID string) (*schemas.BackupSession, error) {
	var s schemas.BackupSession
	if err := r.DB.Where("session_id = ? AND user_id = ?", sessionID, userID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *BackupRepository) GetSessionByKey(sessionID string, keyID uint) (*schemas.BackupSession, error) {
	var s schemas.BackupSession
	// Session must be from today to be still writable
	today := time.Now().Format("2006-01-02")
	if err := r.DB.Where("session_id = ? AND key_id = ? AND date = ?", sessionID, keyID, today).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}
