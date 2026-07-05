package backupServices

import (
	"backend/src/external"
	"backend/src/internal/logger"
	backupRepo "backend/src/internal/repository/backup"
	keyRepo "backend/src/internal/repository/key"
	userRepo "backend/src/internal/repository/user"
	"backend/src/internal/schemas"
	"backend/src/pkg/apperr"
	"backend/src/utils"
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// backupsPerDay defines max backup sessions per day per API key, by user tier.
var backupsPerDay = map[string]int64{
	"free":       1,
	"starter":    2,
	"pro":        3,
	"enterprise": 4,
}

type BackupService struct {
	BackupRepo *backupRepo.BackupRepository
	UserRepo   *userRepo.UserRepository
	KeyRepo    *keyRepo.KeyRepository
	CF         *external.CloudflareClient
}

func New(br *backupRepo.BackupRepository, ur *userRepo.UserRepository, kr *keyRepo.KeyRepository) *BackupService {
	return &BackupService{
		BackupRepo: br,
		UserRepo:   ur,
		KeyRepo:    kr,
		CF:         external.NewCloudflareClient(),
	}
}

// ── Backup bucket ─────────────────────────────────────────────────────────────

// EnsureBackupBucket returns the hidden backup bucket for userID, creating it if needed.
func (s *BackupService) EnsureBackupBucket(ctx context.Context, userID string) (*schemas.BackupBucket, error) {
	b, err := s.BackupRepo.GetBucketByUserID(userID)
	if err == nil {
		return b, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, apperr.Internal("failed to get backup bucket", err)
	}
	return s.createBackupBucket(ctx, userID)
}

func (s *BackupService) createBackupBucket(ctx context.Context, userID string) (*schemas.BackupBucket, error) {
	r2Name := fmt.Sprintf("fkbkp-%s", strings.ToLower(userID))
	if len(r2Name) > 63 {
		r2Name = r2Name[:63]
	}

	if err := s.CF.CreateR2Bucket(ctx, r2Name); err != nil {
		// Ignore already-exists errors
		if !strings.Contains(err.Error(), "already exists") && !strings.Contains(err.Error(), "10006") {
			return nil, apperr.Internal("failed to provision backup bucket", err)
		}
	}

	b := &schemas.BackupBucket{
		UserID: userID,
		R2Name: r2Name,
	}
	if err := s.BackupRepo.CreateBucket(b); err != nil {
		return nil, apperr.Internal("failed to save backup bucket", err)
	}
	logger.Info("backup bucket created", zap.String("user_id", userID), zap.String("r2_name", r2Name))
	return b, nil
}

// ── Upload ────────────────────────────────────────────────────────────────────

type UploadResult struct {
	SessionID  string `json:"session_id"`
	PathPrefix string `json:"path_prefix"`
	Remaining  int64  `json:"remaining_today"`
	Skipped    bool   `json:"skipped,omitempty"`
}

// UploadFile uploads one file to the user's backup bucket.
//
// fileType must be "image" or "db":
//   - "image": stored at images/{path}, skipped if already exists (idempotent).
//   - "db":    stored at db/{session.PathPrefix}/{path}, always a new timestamped folder.
//
// sessionID groups files from the same backup cycle. On first call (sessionID==""),
// quota is checked and a new session is created. Subsequent calls reuse the session.
func (s *BackupService) UploadFile(ctx context.Context, keyID uint, userID, sessionID, filePath, fileType string, data []byte, contentType string) (*UploadResult, error) {
	if fileType != "image" && fileType != "db" {
		fileType = "db"
	}

	user, err := s.UserRepo.GetUserByID(userID)
	if err != nil {
		return nil, apperr.NotFound("user not found")
	}

	today := time.Now().Format("2006-01-02")

	var session *schemas.BackupSession

	if sessionID == "" {
		// New backup cycle — check quota
		quota, ok := backupsPerDay[user.Tier]
		if !ok {
			quota = 1
		}
		used, err := s.BackupRepo.CountTodayByKey(keyID, today)
		if err != nil {
			return nil, apperr.Internal("failed to check quota", err)
		}
		if used >= quota {
			return nil, apperr.TooManyRequests(fmt.Sprintf("limite de %d backup(s) por dia atingido para este plano", quota))
		}

		prefix := time.Now().Format("2006-01-02/15-04-05")
		session = &schemas.BackupSession{
			SessionID:  utils.GenerateULID(),
			KeyID:      keyID,
			UserID:     userID,
			Date:       today,
			PathPrefix: prefix,
		}
		if err := s.BackupRepo.CreateSession(session); err != nil {
			return nil, apperr.Internal("failed to create backup session", err)
		}
	} else {
		session, err = s.BackupRepo.GetSessionByKey(sessionID, keyID)
		if err != nil {
			return nil, apperr.NotFound("sessão de backup não encontrada ou expirada")
		}
	}

	bucket, err := s.EnsureBackupBucket(ctx, userID)
	if err != nil {
		return nil, err
	}

	cleanPath := strings.TrimLeft(filePath, "/")

	var r2Key string
	if fileType == "image" {
		// Images always go to the same flat folder; path is preserved as-is.
		r2Key = "images/" + cleanPath
	} else {
		// DB dumps go under a timestamped subfolder inside db/.
		r2Key = "db/" + session.PathPrefix + "/" + cleanPath
	}

	s3Client, err := s.CF.NewR2S3Client(ctx)
	if err != nil {
		return nil, apperr.Internal("failed to init S3 client", err)
	}

	// For images: skip upload if the object already exists in R2.
	// HeadObject returns no error when the object exists; any error means proceed.
	skipped := false
	if fileType == "image" {
		_, headErr := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(bucket.R2Name),
			Key:    aws.String(r2Key),
		})
		if headErr == nil {
			skipped = true
		}
	}

	if !skipped {
		_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:        aws.String(bucket.R2Name),
			Key:           aws.String(r2Key),
			Body:          bytes.NewReader(data),
			ContentType:   aws.String(contentType),
			ContentLength: aws.Int64(int64(len(data))),
		})
		if err != nil {
			return nil, apperr.Internal("upload to backup bucket failed", err)
		}
		s.BackupRepo.IncrStats(session.SessionID, 1, int64(len(data)))
	}

	quota := backupsPerDay[user.Tier]
	if quota == 0 {
		quota = 1
	}
	used, _ := s.BackupRepo.CountTodayByKey(keyID, today)
	remaining := quota - used

	return &UploadResult{
		SessionID:  session.SessionID,
		PathPrefix: session.PathPrefix,
		Remaining:  remaining,
		Skipped:    skipped,
	}, nil
}

// ── Quota ─────────────────────────────────────────────────────────────────────

type QuotaResult struct {
	Used      int64  `json:"used"`
	Max       int64  `json:"max"`
	Remaining int64  `json:"remaining"`
	Date      string `json:"date"`
}

func (s *BackupService) GetQuota(keyID uint, userID string) (*QuotaResult, error) {
	user, err := s.UserRepo.GetUserByID(userID)
	if err != nil {
		return nil, apperr.NotFound("user not found")
	}
	today := time.Now().Format("2006-01-02")
	max, ok := backupsPerDay[user.Tier]
	if !ok {
		max = 1
	}
	used, err := s.BackupRepo.CountTodayByKey(keyID, today)
	if err != nil {
		return nil, apperr.Internal("failed to check quota", err)
	}
	remaining := max - used
	if remaining < 0 {
		remaining = 0
	}
	return &QuotaResult{Used: used, Max: max, Remaining: remaining, Date: today}, nil
}

// ── List sessions (for UI) ────────────────────────────────────────────────────

type SessionResponse struct {
	SessionID  string `json:"session_id"`
	Date       string `json:"date"`
	PathPrefix string `json:"path_prefix"`
	FileCount  int    `json:"file_count"`
	TotalSize  int64  `json:"total_size"`
	CreatedAt  string `json:"created_at"`
}

func (s *BackupService) ListSessions(userID, date string, page, limit int) ([]SessionResponse, int64, error) {
	sessions, total, err := s.BackupRepo.ListByUserID(userID, date, page, limit)
	if err != nil {
		return nil, 0, apperr.Internal("failed to list sessions", err)
	}
	result := make([]SessionResponse, 0, len(sessions))
	for _, sess := range sessions {
		r := SessionResponse{
			SessionID:  sess.SessionID,
			Date:       sess.Date,
			PathPrefix: sess.PathPrefix,
			FileCount:  sess.FileCount,
			TotalSize:  sess.TotalSize,
		}
		if sess.CreatedAt != nil {
			r.CreatedAt = sess.CreatedAt.Format("2006-01-02 15:04:05")
		}
		result = append(result, r)
	}
	return result, total, nil
}
