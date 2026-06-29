package storageConfigRepo

import (
	"backend/src/internal/schemas"

	"gorm.io/gorm"
)

type BucketRepository struct {
	DB *gorm.DB
}

func NewStorageConfigRepository(db *gorm.DB) *BucketRepository {
	return &BucketRepository{DB: db}
}

func (r *BucketRepository) Create(b *schemas.Bucket) error {
	return r.DB.Create(b).Error
}

func (r *BucketRepository) ListByUserID(userID string) ([]schemas.Bucket, error) {
	var buckets []schemas.Bucket
	err := r.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&buckets).Error
	return buckets, err
}

func (r *BucketRepository) GetByID(bucketID, userID string) (*schemas.Bucket, error) {
	var b schemas.Bucket
	if err := r.DB.Where("bucket_id = ? AND user_id = ?", bucketID, userID).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BucketRepository) SetStatus(bucketID string, status schemas.BucketStatus) error {
	return r.DB.Model(&schemas.Bucket{}).Where("bucket_id = ?", bucketID).Update("status", status).Error
}

func (r *BucketRepository) Delete(bucketID, userID string) error {
	return r.DB.Where("bucket_id = ? AND user_id = ?", bucketID, userID).Delete(&schemas.Bucket{}).Error
}

func (r *BucketRepository) CountByUserID(userID string) (int64, error) {
	var n int64
	r.DB.Model(&schemas.Bucket{}).Where("user_id = ?", userID).Count(&n)
	return n, nil
}

func (r *BucketRepository) IncrStats(bucketID string, deltaFiles, deltaBytes int64) {
	r.DB.Model(&schemas.Bucket{}).Where("bucket_id = ?", bucketID).
		Updates(map[string]interface{}{
			"file_count": gorm.Expr("file_count + ?", deltaFiles),
			"bytes_used": gorm.Expr("bytes_used + ?", deltaBytes),
		})
}

type BucketStats struct {
	TotalFiles int64
	BytesUsed  int64
}

func (r *BucketRepository) SumStatsByUserID(userID string) (*BucketStats, error) {
	var res BucketStats
	err := r.DB.Model(&schemas.Bucket{}).
		Select("COALESCE(SUM(file_count),0) as total_files, COALESCE(SUM(bytes_used),0) as bytes_used").
		Where("user_id = ?", userID).
		Scan(&res).Error
	return &res, err
}

func (r *BucketRepository) SetPublicDomain(bucketID, domain string) error {
	return r.DB.Model(&schemas.Bucket{}).Where("bucket_id = ?", bucketID).
		Updates(map[string]any{"public_domain": domain, "public_access_enabled": true}).Error
}

func (r *BucketRepository) SetPublicAccessEnabled(bucketID string, enabled bool) error {
	return r.DB.Model(&schemas.Bucket{}).Where("bucket_id = ?", bucketID).
		Update("public_access_enabled", enabled).Error
}

func (r *BucketRepository) SetCustomDomain(bucketID, domain string) error {
	return r.DB.Model(&schemas.Bucket{}).Where("bucket_id = ?", bucketID).
		Update("custom_domain", domain).Error
}

func (r *BucketRepository) ListActiveWithoutPublicAccess() ([]schemas.Bucket, error) {
	var buckets []schemas.Bucket
	err := r.DB.Where("status = ? AND (public_access_enabled = false OR public_domain = '')", schemas.BucketStatusActive).Find(&buckets).Error
	return buckets, err
}

// ListActiveWithoutPublicURL retorna buckets ativos que não têm nenhum domínio público configurado.
func (r *BucketRepository) ListActiveWithoutCustomDomain() ([]schemas.Bucket, error) {
	var buckets []schemas.Bucket
	err := r.DB.Where(
		"status = ? AND (custom_domain IS NULL OR custom_domain = '') AND (public_domain IS NULL OR public_domain = '')",
		schemas.BucketStatusActive,
	).Find(&buckets).Error
	return buckets, err
}

func (r *BucketRepository) CountAll() (int64, error) {
	var n int64
	return n, r.DB.Model(&schemas.Bucket{}).Count(&n).Error
}

func (r *BucketRepository) CountByStatus(status string) (int64, error) {
	var n int64
	return n, r.DB.Model(&schemas.Bucket{}).Where("status = ?", status).Count(&n).Error
}
