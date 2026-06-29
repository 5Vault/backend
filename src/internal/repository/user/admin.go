package user

import (
	"backend/src/internal/schemas"
)

func (repo *UserRepository) ListUsers(page, limit int, search, tier string) ([]schemas.User, int64, error) {
	var users []schemas.User
	var total int64

	q := repo.DB.Model(&schemas.User{})

	if search != "" {
		like := "%" + search + "%"
		q = q.Where("username LIKE ? OR email LIKE ? OR name LIKE ?", like, like, like)
	}
	if tier != "" {
		q = q.Where("tier = ?", tier)
	}

	q.Count(&total)

	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (repo *UserRepository) SetUserTier(userID, tier string) error {
	return repo.DB.Model(&schemas.User{}).Where("user_id = ?", userID).Update("tier", tier).Error
}

func (repo *UserRepository) SetUserRole(userID, role string) error {
	return repo.DB.Model(&schemas.User{}).Where("user_id = ?", userID).Update("role", role).Error
}

func (repo *UserRepository) HardDeleteUser(userID string) error {
	return repo.DB.Unscoped().Where("user_id = ?", userID).Delete(&schemas.User{}).Error
}

func (repo *UserRepository) CountByTier() (map[string]int64, error) {
	type row struct {
		Tier  string
		Count int64
	}
	var rows []row
	if err := repo.DB.Model(&schemas.User{}).Select("tier, count(*) as count").Group("tier").Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, r := range rows {
		out[r.Tier] = r.Count
	}
	return out, nil
}

func (repo *UserRepository) TotalUsers() (int64, error) {
	var count int64
	return count, repo.DB.Model(&schemas.User{}).Count(&count).Error
}

func (repo *UserRepository) NewUsersThisMonth() (int64, error) {
	var count int64
	return count, repo.DB.Model(&schemas.User{}).
		Where("created_at >= DATE_FORMAT(NOW(), '%Y-%m-01')").
		Count(&count).Error
}
