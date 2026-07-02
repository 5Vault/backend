package bucketServices

import (
	"backend/src/external"
	"backend/src/internal/logger"
	"backend/src/internal/models"
	fileRepo "backend/src/internal/repository/file"
	dirRepo "backend/src/internal/repository/storage"
	bucketRepo "backend/src/internal/repository/storage_config"
	userRepo "backend/src/internal/repository/user"
	"backend/src/internal/schemas"
	"backend/src/pkg/apperr"
	"backend/src/utils"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// -1 = ilimitado
var tierBucketLimit = map[string]int64{
	"free":       1,
	"starter":    3,
	"pro":        6,
	"enterprise": -1,
}

const fileCacheTTL = 2 * time.Minute

type BucketService struct {
	BucketRepo *bucketRepo.BucketRepository
	DirRepo    *dirRepo.DirectoryRepository
	UserRepo   *userRepo.UserRepository
	FileRepo   *fileRepo.StorageRepository
	CF         *external.CloudflareClient
	Redis      *redis.Client
}

func NewBucketService(br *bucketRepo.BucketRepository, dr *dirRepo.DirectoryRepository, ur *userRepo.UserRepository, fr *fileRepo.StorageRepository, rdb *redis.Client) *BucketService {
	return &BucketService{BucketRepo: br, DirRepo: dr, UserRepo: ur, FileRepo: fr, CF: external.NewCloudflareClient(), Redis: rdb}
}

func (s *BucketService) cacheKey(bucketID, dirID string, page, limit int) string {
	return fmt.Sprintf("files:%s:%s:%d:%d", bucketID, dirID, page, limit)
}

func (s *BucketService) invalidateCache(ctx context.Context, bucketID, dirID string) {
	pattern := fmt.Sprintf("files:%s:%s:*", bucketID, dirID)
	keys, err := s.Redis.Keys(ctx, pattern).Result()
	if err != nil || len(keys) == 0 {
		return
	}
	s.Redis.Del(ctx, keys...)
}

// ── Buckets ───────────────────────────────────────────────────────────────────

func (s *BucketService) CreateBucket(userID, name string) (*models.ResponseBucket, error) {
	user, err := s.UserRepo.GetUserByID(userID)
	if err != nil {
		return nil, apperr.NotFound("user not found")
	}
	limit, ok := tierBucketLimit[user.Tier]
	if !ok {
		limit = 1
	}
	if limit >= 0 {
		count, err := s.BucketRepo.CountByUserID(userID)
		if err != nil {
			return nil, apperr.Internal("failed to count buckets", err)
		}
		if count >= limit {
			return nil, apperr.Forbidden(fmt.Sprintf("seu plano permite no máximo %d bucket(s)", limit))
		}
	}

	bucketID := utils.GenerateULID()
	r2Name := fmt.Sprintf("fv-%s", bucketID)

	b := &schemas.Bucket{
		BucketID: bucketID,
		UserID:   userID,
		Name:     name,
		R2Name:   r2Name,
		Status:   schemas.BucketStatusPending,
	}
	if err := s.BucketRepo.Create(b); err != nil {
		return nil, apperr.Internal("failed to create bucket", err)
	}

	go s.provision(context.Background(), b)

	return toBucketResponse(b), nil
}

func (s *BucketService) ListBuckets(userID string) ([]models.ResponseBucket, error) {
	buckets, err := s.BucketRepo.ListByUserID(userID)
	if err != nil {
		return nil, apperr.Internal("failed to list buckets", err)
	}
	result := make([]models.ResponseBucket, 0, len(buckets))
	for i := range buckets {
		result = append(result, *toBucketResponse(&buckets[i]))
	}
	return result, nil
}

func (s *BucketService) DeleteBucket(bucketID, userID string) error {
	b, err := s.BucketRepo.GetByID(bucketID, userID)
	if err != nil {
		return apperr.NotFound("bucket not found")
	}
	// Esvazia e remove o bucket R2 em background; mesmo se falhar, remove o registro local
	go func() {
		ctx := context.Background()
		if err := s.CF.EmptyAndDeleteR2Bucket(ctx, b.R2Name); err != nil {
			logger.Warn("failed to empty/delete R2 bucket", zap.String("r2_name", b.R2Name), zap.Error(err))
		}
	}()
	s.FileRepo.DeleteByStorageID(bucketID, userID)
	return s.BucketRepo.Delete(bucketID, userID)
}

// ── Diretórios ────────────────────────────────────────────────────────────────

func (s *BucketService) CreateDirectory(bucketID, userID, name string) (*models.ResponseDirectory, error) {
	if _, err := s.BucketRepo.GetByID(bucketID, userID); err != nil {
		return nil, apperr.NotFound("bucket not found")
	}
	d := &schemas.Directory{
		DirID:    utils.GenerateULID(),
		BucketID: bucketID,
		UserID:   userID,
		Name:     name,
	}
	if err := s.DirRepo.Create(d); err != nil {
		return nil, apperr.Internal("failed to create directory", err)
	}
	return toDirResponse(d), nil
}

func (s *BucketService) ListDirectories(bucketID, userID string) ([]models.ResponseDirectory, error) {
	if _, err := s.BucketRepo.GetByID(bucketID, userID); err != nil {
		return nil, apperr.NotFound("bucket not found")
	}
	dirs, err := s.DirRepo.ListByBucket(bucketID, userID)
	if err != nil {
		return nil, apperr.Internal("failed to list directories", err)
	}
	result := make([]models.ResponseDirectory, 0, len(dirs))
	for i := range dirs {
		result = append(result, *toDirResponse(&dirs[i]))
	}
	return result, nil
}

func (s *BucketService) DeleteDirectory(dirID, bucketID, userID string) error {
	if _, err := s.DirRepo.GetByID(dirID, bucketID, userID); err != nil {
		return apperr.NotFound("directory not found")
	}
	return s.DirRepo.Delete(dirID, bucketID, userID)
}

// ── Arquivos ──────────────────────────────────────────────────────────────────

func (s *BucketService) UploadFile(ctx context.Context, bucketID, dirID, userID, filename string, body []byte, contentType string) (*models.ResponseUploadFile, error) {
	b, err := s.BucketRepo.GetByID(bucketID, userID)
	if err != nil {
		return nil, apperr.NotFound("bucket not found")
	}
	if b.Status != schemas.BucketStatusActive {
		return nil, apperr.BadRequest("bucket não está ativo ainda")
	}

	var key string
	if dirID == "root" {
		key = filename
	} else {
		dir, err := s.DirRepo.GetByID(dirID, bucketID, userID)
		if err != nil {
			return nil, apperr.NotFound("directory not found")
		}
		key = fmt.Sprintf("%s/%s", dir.DirID, filename)
	}
	s3Client, err := s.CF.NewR2S3Client(ctx)
	if err != nil {
		return nil, apperr.Internal("failed to init S3 client", err)
	}

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(b.R2Name),
		Key:           aws.String(key),
		Body:          strings.NewReader(string(body)),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(int64(len(body))),
	})
	if err != nil {
		return nil, apperr.Internal("upload failed", err)
	}

	s.invalidateCache(ctx, bucketID, dirID)
	s.BucketRepo.IncrStats(bucketID, 1, int64(len(body)))

	publicURL := r2PublicURL(b, key)
	_, err = s.FileRepo.CreateFile(&schemas.File{
		FileID:     utils.GenerateULID(),
		UserID:     userID,
		StorageID:  bucketID,
		FileType:   contentType,
		FileURL:    publicURL,
		FileSize:   int64(len(body)),
		UploadedAt: time.Now(),
	})
	if err != nil {
		return nil, err
	}

	return &models.ResponseUploadFile{
		FileName:  filename,
		PublicURL: publicURL,
		Size:      int64(len(body)),
	}, nil
}

func (s *BucketService) UploadFilePublic(ctx context.Context, bucketID, userID, filename string, body []byte, contentType string, forceCreate bool) (*models.ResponseUploadFile, error) {
	b, err := s.BucketRepo.GetByID(bucketID, userID)
	if err != nil {
		return nil, apperr.NotFound("bucket not found")
	}
	if b.Status != schemas.BucketStatusActive {
		return nil, apperr.BadRequest("bucket não está ativo ainda")
	}

	var key string
	var dirPath, baseName string
	lastSlash := strings.LastIndex(filename, "/")
	if lastSlash != -1 {
		dirPath = filename[:lastSlash]
		baseName = filename[lastSlash+1:]
	} else {
		baseName = filename
	}

	var dirID string = "root"

	if dirPath != "" {
		dir, err := s.DirRepo.GetByName(dirPath, bucketID, userID)
		if err != nil {
			if forceCreate {
				newDir := &schemas.Directory{
					DirID:    utils.GenerateULID(),
					BucketID: bucketID,
					UserID:   userID,
					Name:     dirPath,
				}
				if err := s.DirRepo.Create(newDir); err != nil {
					return nil, apperr.Internal("failed to create directory", err)
				}
				dirID = newDir.DirID
			} else {
				return nil, apperr.BadRequest(fmt.Sprintf("diretório '%s' não existe no bucket", dirPath))
			}
		} else {
			dirID = dir.DirID
		}
	}

	if dirID == "root" {
		key = baseName
	} else {
		key = fmt.Sprintf("%s/%s", dirID, baseName)
	}

	s3Client, err := s.CF.NewR2S3Client(ctx)
	if err != nil {
		return nil, apperr.Internal("failed to init S3 client", err)
	}

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(b.R2Name),
		Key:           aws.String(key),
		Body:          strings.NewReader(string(body)),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(int64(len(body))),
	})
	if err != nil {
		return nil, apperr.Internal("upload failed", err)
	}

	s.invalidateCache(ctx, bucketID, dirID)
	s.BucketRepo.IncrStats(bucketID, 1, int64(len(body)))

	publicURLPublic := r2PublicURL(b, key)
	_, err = s.FileRepo.CreateFile(&schemas.File{
		FileID:     utils.GenerateULID(),
		UserID:     userID,
		StorageID:  bucketID,
		FileType:   contentType,
		FileURL:    publicURLPublic,
		FileSize:   int64(len(body)),
		UploadedAt: time.Now(),
	})
	if err != nil {
		return nil, err
	}

	return &models.ResponseUploadFile{
		FileName:  baseName,
		PublicURL: publicURLPublic,
		Size:      int64(len(body)),
	}, nil
}

func (s *BucketService) ListFiles(ctx context.Context, bucketID, dirID, userID string, page, limit int) (*models.ResponseListFiles, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// cache hit
	cacheKey := s.cacheKey(bucketID, dirID, page, limit)
	if cached, err := s.Redis.Get(ctx, cacheKey).Bytes(); err == nil {
		var res models.ResponseListFiles
		if json.Unmarshal(cached, &res) == nil {
			return &res, nil
		}
	}

	b, err := s.BucketRepo.GetByID(bucketID, userID)
	if err != nil {
		return nil, apperr.NotFound("bucket not found")
	}
	if dirID != "root" {
		if _, err := s.DirRepo.GetByID(dirID, bucketID, userID); err != nil {
			return nil, apperr.NotFound("directory not found")
		}
	}

	s3Client, err := s.CF.NewR2S3Client(ctx)
	if err != nil {
		return nil, apperr.Internal("failed to init S3 client", err)
	}

	var prefix string
	if dirID != "root" {
		prefix = dirID + "/"
	}

	// busca tudo com paginação via ContinuationToken
	var allFiles []models.FileEntry
	var token *string
	for {
		out, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(b.R2Name),
			Prefix:            aws.String(prefix),
			Delimiter:         aws.String("/"),
			ContinuationToken: token,
		})
		if err != nil {
			return nil, apperr.Internal("list failed", err)
		}
		for _, obj := range out.Contents {
			name := strings.TrimPrefix(aws.ToString(obj.Key), prefix)
			if name == "" {
				continue
			}
			allFiles = append(allFiles, models.FileEntry{
				Key:          aws.ToString(obj.Key),
				Size:         aws.ToInt64(obj.Size),
				LastModified: obj.LastModified.Format("2006-01-02T15:04:05Z"),
				PublicURL:    r2PublicURL(b, aws.ToString(obj.Key)),
			})
		}
		if !aws.ToBool(out.IsTruncated) {
			break
		}
		token = out.NextContinuationToken
	}

	total := len(allFiles)
	start := (page - 1) * limit
	end := start + limit
	if start >= total {
		allFiles = []models.FileEntry{}
	} else {
		if end > total {
			end = total
		}
		allFiles = allFiles[start:end]
	}

	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}
	res := &models.ResponseListFiles{Files: allFiles, Total: total, Page: page, Limit: limit, TotalPages: totalPages}

	// armazena no cache
	if data, err := json.Marshal(res); err == nil {
		s.Redis.Set(ctx, cacheKey, data, fileCacheTTL)
	}

	return res, nil
}

func (s *BucketService) DeleteFile(ctx context.Context, bucketID, dirID, userID, filename string) error {
	b, err := s.BucketRepo.GetByID(bucketID, userID)
	if err != nil {
		return apperr.NotFound("bucket not found")
	}
	if dirID != "root" {
		if _, err := s.DirRepo.GetByID(dirID, bucketID, userID); err != nil {
			return apperr.NotFound("directory not found")
		}
	}

	s3Client, err := s.CF.NewR2S3Client(ctx)
	if err != nil {
		return apperr.Internal("failed to init S3 client", err)
	}

	var key string
	if dirID == "root" {
		key = filename
	} else {
		key = fmt.Sprintf("%s/%s", dirID, filename)
	}

	fileURL := r2PublicURL(b, key)
	fileRecord, _ := s.FileRepo.GetByURL(fileURL, userID)

	_, err = s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(b.R2Name),
		Key:    aws.String(key),
	})
	if err == nil {
		s.invalidateCache(ctx, bucketID, dirID)
		var deltaBytes int64
		if fileRecord != nil {
			deltaBytes = -fileRecord.FileSize
		}
		s.BucketRepo.IncrStats(bucketID, -1, deltaBytes)
		s.FileRepo.DeleteByURL(fileURL, userID)
	}
	return err
}

func (s *BucketService) ListFilesPublic(ctx context.Context, bucketID, userID string, dirName string, page, limit int) (*models.ResponseListFiles, error) {
	b, err := s.BucketRepo.GetByID(bucketID, userID)
	if err != nil {
		return nil, apperr.NotFound("bucket not found")
	}
	if b.Status != schemas.BucketStatusActive {
		return nil, apperr.BadRequest("bucket não está ativo ainda")
	}

	var dirID string = "root"
	if dirName != "" {
		dir, err := s.DirRepo.GetByName(dirName, bucketID, userID)
		if err != nil {
			return nil, apperr.NotFound("directory not found")
		}
		dirID = dir.DirID
	}

	return s.ListFiles(ctx, bucketID, dirID, userID, page, limit)
}

func (s *BucketService) DeleteFilePublic(ctx context.Context, bucketID, userID, filename string) error {
	b, err := s.BucketRepo.GetByID(bucketID, userID)
	if err != nil {
		return apperr.NotFound("bucket not found")
	}
	if b.Status != schemas.BucketStatusActive {
		return apperr.BadRequest("bucket não está ativo ainda")
	}

	var dirPath, baseName string
	lastSlash := strings.LastIndex(filename, "/")
	if lastSlash != -1 {
		dirPath = filename[:lastSlash]
		baseName = filename[lastSlash+1:]
	} else {
		baseName = filename
	}

	var dirID string = "root"
	if dirPath != "" {
		dir, err := s.DirRepo.GetByName(dirPath, bucketID, userID)
		if err != nil {
			return apperr.NotFound("directory not found")
		}
		dirID = dir.DirID
	}

	return s.DeleteFile(ctx, bucketID, dirID, userID, baseName)
}

func (s *BucketService) GetStats(userID string) (*models.BucketStats, error) {
	res, err := s.BucketRepo.SumStatsByUserID(userID)
	if err != nil {
		return nil, apperr.Internal("failed to get bucket stats", err)
	}
	return &models.BucketStats{TotalFiles: res.TotalFiles, BytesUsed: res.BytesUsed}, nil
}

func r2PublicURL(b *schemas.Bucket, key string) string {
	if b.CustomDomain != "" {
		return fmt.Sprintf("https://%s/%s", strings.TrimRight(b.CustomDomain, "/"), key)
	}
	if b.PublicDomain != "" {
		return fmt.Sprintf("https://%s/%s/%s", strings.TrimRight(b.PublicDomain, "/"), b.R2Name, key)
	}
	base := strings.TrimRight(os.Getenv("R2_PUBLIC_BASE"), "/")
	if base == "" {
		return "/" + key
	}
	return fmt.Sprintf("%s/%s/%s", base, b.R2Name, key)
}

// ── interno ───────────────────────────────────────────────────────────────────

// storageDomain retorna o domínio base do storage (ex: "5vault.app").
func storageDomain() string { return os.Getenv("STORAGE_DOMAIN") }

// defaultSubdomain retorna o subdomínio padrão de um bucket: "{bucketID}.sexdaily.app".
func defaultSubdomain(bucketID string) string {
	d := storageDomain()
	if d == "" {
		return ""
	}
	return bucketID + "." + d
}

// SetCustomDomain configura um subdomínio *.5vault.app personalizado para o bucket.
// Apenas usuários pro/enterprise podem alterar o subdomínio padrão.
func (s *BucketService) SetCustomDomain(ctx context.Context, bucketID, userID, subdomain string) error {
	b, err := s.BucketRepo.GetByID(bucketID, userID)
	if err != nil {
		return apperr.NotFound("bucket not found")
	}

	d := storageDomain()
	if d != "" && subdomain != "" {
		if !strings.HasSuffix(subdomain, "."+d) {
			return apperr.BadRequest(fmt.Sprintf("o subdomínio deve terminar com .%s", d))
		}
		// apenas pro/enterprise podem personalizar
		user, err := s.UserRepo.GetUserByID(userID)
		if err != nil {
			return apperr.Internal("failed to get user", err)
		}
		if user.Tier != "pro" && user.Tier != "enterprise" {
			return apperr.Forbidden("subdomínio personalizado disponível apenas nos planos Pro e Enterprise")
		}
	}

	if subdomain != "" {
		if err := s.CF.AttachCustomDomain(ctx, b.R2Name, subdomain); err != nil {
			return apperr.Internal("failed to attach custom domain", err)
		}
	}
	return s.BucketRepo.SetCustomDomain(bucketID, subdomain)
}

// EnablePublicAccess ativa acesso público para um bucket específico e salva o domínio.
func (s *BucketService) EnablePublicAccess(ctx context.Context, bucketID, userID string) (string, error) {
	b, err := s.BucketRepo.GetByID(bucketID, userID)
	if err != nil {
		return "", apperr.NotFound("bucket not found")
	}
	domain, err := s.CF.AllowPublicAccess(ctx, b.R2Name)
	if err != nil {
		return "", apperr.Internal("failed to enable public access", err)
	}
	_ = s.BucketRepo.SetPublicDomain(b.BucketID, domain)
	return domain, nil
}

// EnsureDefaultDomain garante que todos os buckets ativos tenham um domínio público.
// Se STORAGE_DOMAIN estiver configurado, anexa "{userID}.{STORAGE_DOMAIN}" via custom domain.
// Caso contrário, usa o domínio gerenciado pub-*.r2.dev como fallback.
func (s *BucketService) EnsureDefaultDomain(ctx context.Context) {
	buckets, err := s.BucketRepo.ListActiveWithoutCustomDomain()
	if err != nil || len(buckets) == 0 {
		return
	}

	sd := storageDomain()

	for _, b := range buckets {
		if sd != "" {
			sub := b.BucketID + "." + sd
			if err := s.CF.AttachCustomDomain(ctx, b.R2Name, sub); err != nil {
				logger.Warn("failed to attach default domain", zap.String("bucket", b.R2Name), zap.String("domain", sub), zap.Error(err))
			} else {
				_ = s.BucketRepo.SetCustomDomain(b.BucketID, sub)
				logger.Info("default domain attached", zap.String("bucket", b.R2Name), zap.String("domain", sub))
				continue
			}
		}
		// fallback: domínio gerenciado pub-*.r2.dev
		if b.PublicDomain == "" {
			domain, err := s.CF.AllowPublicAccess(ctx, b.R2Name)
			if err != nil {
				logger.Warn("failed to enable managed public access", zap.String("bucket", b.R2Name), zap.Error(err))
				continue
			}
			_ = s.BucketRepo.SetPublicDomain(b.BucketID, domain)
			logger.Info("managed public access enabled", zap.String("bucket", b.R2Name), zap.String("domain", domain))
		}
	}
}

func (s *BucketService) provision(ctx context.Context, b *schemas.Bucket) {
	log := logger.With(zap.String("bucket_id", b.BucketID), zap.String("r2_name", b.R2Name))
	log.Info("provisioning bucket")

	if err := s.CF.CreateR2Bucket(ctx, b.R2Name); err != nil {
		log.Error("failed to create R2 bucket", zap.Error(err))
		_ = s.BucketRepo.SetStatus(b.BucketID, schemas.BucketStatusError)
		return
	}

	// Tenta anexar o subdomínio padrão do bucket; cai no pub-*.r2.dev como fallback.
	if sub := defaultSubdomain(b.BucketID); sub != "" {
		if err := s.CF.AttachCustomDomain(ctx, b.R2Name, sub); err != nil {
			log.Warn("could not attach default domain, falling back to managed", zap.Error(err))
			if domain, err := s.CF.AllowPublicAccess(ctx, b.R2Name); err != nil {
				log.Warn("could not enable managed public access", zap.Error(err))
			} else {
				_ = s.BucketRepo.SetPublicDomain(b.BucketID, domain)
			}
		} else {
			_ = s.BucketRepo.SetCustomDomain(b.BucketID, sub)
		}
	} else {
		if domain, err := s.CF.AllowPublicAccess(ctx, b.R2Name); err != nil {
			log.Warn("could not enable managed public access", zap.Error(err))
		} else {
			_ = s.BucketRepo.SetPublicDomain(b.BucketID, domain)
		}
	}

	_ = s.BucketRepo.SetStatus(b.BucketID, schemas.BucketStatusActive)
	log.Info("bucket active")
}

func toBucketResponse(b *schemas.Bucket) *models.ResponseBucket {
	createdAt := ""
	if b.CreatedAt != nil {
		createdAt = b.CreatedAt.Format("2006-01-02T15:04:05Z")
	}
	return &models.ResponseBucket{
		BucketID:            b.BucketID,
		UserID:              b.UserID,
		Name:                b.Name,
		R2Name:              b.R2Name,
		Status:              string(b.Status),
		CustomDomain:        b.CustomDomain,
		PublicDomain:        b.PublicDomain,
		PublicAccessEnabled: b.PublicAccessEnabled,
		CreatedAt:           createdAt,
	}
}

func toDirResponse(d *schemas.Directory) *models.ResponseDirectory {
	createdAt := ""
	if d.CreatedAt != nil {
		createdAt = d.CreatedAt.Format("2006-01-02T15:04:05Z")
	}
	return &models.ResponseDirectory{
		DirID:     d.DirID,
		BucketID:  d.BucketID,
		Name:      d.Name,
		CreatedAt: createdAt,
	}
}
