package repositories

import (
	"backend/internal/schemas"

	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (repo *UserRepository) CreateUser(user *schemas.User) error {
	if err := repo.DB.Create(user).Error; err != nil {
		return err
	}
	return nil
}

func (repo *UserRepository) GetUser(id string) (*schemas.User, error) {
	var user *schemas.User
	if err := repo.DB.First(&user, "user_id = ? OR email = ? or username = ? or phone = ?", id, id, id, id).Error; err != nil {
		return nil, err // Other error
	}
	return user, nil
}

func (repo *UserRepository) GetUserByID(id string) (*schemas.User, error) {
	var user *schemas.User
	if err := repo.DB.First(&user, "user_id = ?", id).Error; err != nil {
		return nil, err // Other error
	}
	return user, nil
}

func (repo *UserRepository) GetUserByUsername(username string) (*schemas.User, error) {
	var user *schemas.User
	if err := repo.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *UserRepository) GetUserByEmail(email string) (*schemas.User, error) {
	var user *schemas.User
	if err := repo.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *UserRepository) GetUserByPhone(phone string) (*schemas.User, error) {
	var user *schemas.User
	if err := repo.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *UserRepository) UpdateUser(user *schemas.User) error {
	if err := repo.DB.Save(user).Error; err != nil {
		return err
	}
	return nil
}

func (repo *UserRepository) DeleteUser(id uint) error {
	var user schemas.User
	if err := repo.DB.First(&user, id).Error; err != nil {
		return err
	}
	if err := repo.DB.Delete(&user).Error; err != nil {
		return err
	}
	return nil
}
