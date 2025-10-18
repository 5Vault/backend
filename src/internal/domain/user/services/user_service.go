package services

import (
	keyRepo "backend/src/internal/repository/key"
	usrRepo "backend/src/internal/repository/user"
	"time"

	"backend/src/internal/models"
	"backend/src/internal/schemas"
	utils2 "backend/src/utils"
	"errors"
	"fmt"
	"strings"

	tierService "backend/src/internal/domain/user/tier/service"

	"gorm.io/gorm"
)

var tierSvc = tierService.NewTierService()

type UserService struct {
	UserRepo *usrRepo.UserRepository
	KeyRepo  *keyRepo.KeyRepository
}

func NewUserService(userRepo *usrRepo.UserRepository, keyRepo *keyRepo.KeyRepository) *UserService {
	return &UserService{
		UserRepo: userRepo,
		KeyRepo:  keyRepo,
	}
}

var cSvc = utils2.NewCryptService()

func (s *UserService) CreateUser(user *models.RequestUser) (*string, error) {

	var hashPassword string

	hashPassword = cSvc.HashPassword(user.Password)
	newUserID := utils2.GenerateRandomID()

	dbUser := &schemas.User{
		UserID:        newUserID,
		Username:      user.Username,
		Name:          user.Name,
		Email:         user.Email,
		Password:      hashPassword,
		Phone:         user.Phone,
		Tier:          "free",
		TierUpdatedAt: time.Now(),
	}

	existingUser, err := s.UserRepo.GetUserByID(user.Username)
	if err == nil && existingUser != nil {
		return nil, errors.New("phone already exists")
	}

	existingUser, err = s.UserRepo.GetUserByEmail(user.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("phone already exists")
	}

	existingUser, err = s.UserRepo.GetUserByPhone(user.Phone)
	if err == nil && existingUser != nil {
		return nil, errors.New("phone already exists")
	}

	if err := s.UserRepo.CreateUser(dbUser); err != nil {
		// Verificar tipos específicos de erro do GORM
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.New("user already exists")
		}
		// Verificar erro de violação de constraint por string
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "UNIQUE constraint") ||
			strings.Contains(err.Error(), "violates unique constraint") {
			return nil, errors.New("user already exists")
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	key, err := utils2.GenerateAPIKey()
	if err != nil {
		return nil, err
	}

	keySchema := &schemas.Key{
		Key:    key,
		UserID: newUserID,
	}

	if err := s.KeyRepo.New(keySchema); err != nil {
		return nil, err
	}

	return &newUserID, nil
}

func (s *UserService) GetUserByID(id string, own bool) (*models.ResponseUser, error) {
	user, err := s.UserRepo.GetUser(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	userResponse := models.ResponseUser{
		UserID:        user.UserID,
		Username:      user.Username,
		Name:          user.Name,
		Email:         user.Email,
		Phone:         user.Phone,
		Tier:          user.Tier,
		TierName:      tierSvc.GetTierNameByID(user.Tier),
		TierUpdatedAt: user.TierUpdatedAt.Format("2006-01-02 15:04:05"),
		CreatedAt:     user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     user.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if own {
		key, err := s.KeyRepo.GetByUserID(id)
		if err != nil {
			return nil, fmt.Errorf("error retrieving key: %w", err)
		}
		if key != nil {
			userResponse.ApiKey = &key.Key
		} else {
			userResponse.ApiKey = nil
		}
	}

	return &userResponse, nil
}
