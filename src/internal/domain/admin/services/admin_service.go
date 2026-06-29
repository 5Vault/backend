package adminServices

import (
	"backend/src/internal/logger"
	bucketRepo "backend/src/internal/repository/storage_config"
	usrRepo "backend/src/internal/repository/user"

	"go.uber.org/zap"
)

type AdminService struct {
	UserRepo   *usrRepo.UserRepository
	BucketRepo *bucketRepo.BucketRepository
}

func NewAdminService(userRepo *usrRepo.UserRepository, bucketRepo *bucketRepo.BucketRepository) *AdminService {
	return &AdminService{UserRepo: userRepo, BucketRepo: bucketRepo}
}

type StatsResponse struct {
	TotalUsers      int64            `json:"total_users"`
	NewUsersMonth   int64            `json:"new_users_month"`
	UsersByTier     map[string]int64 `json:"users_by_tier"`
	TotalBuckets    int64            `json:"total_buckets"`
	ActiveBuckets   int64            `json:"active_buckets"`
	PendingBuckets  int64            `json:"pending_buckets"`
}

func (s *AdminService) GetStats() (*StatsResponse, error) {
	total, err := s.UserRepo.TotalUsers()
	if err != nil {
		logger.Error("admin: failed to count users", zap.Error(err))
		return nil, err
	}

	newMonth, _ := s.UserRepo.NewUsersThisMonth()
	byTier, _ := s.UserRepo.CountByTier()
	totalB, _ := s.BucketRepo.CountAll()
	activeB, _ := s.BucketRepo.CountByStatus("active")
	pendingB, _ := s.BucketRepo.CountByStatus("pending")

	return &StatsResponse{
		TotalUsers:     total,
		NewUsersMonth:  newMonth,
		UsersByTier:    byTier,
		TotalBuckets:   totalB,
		ActiveBuckets:  activeB,
		PendingBuckets: pendingB,
	}, nil
}
