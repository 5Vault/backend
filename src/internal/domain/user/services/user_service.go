package services

import (
	lServices "backend/src/internal/domain/login/services"
	"backend/src/internal/domain/user/repositories"
	"backend/src/internal/models"
	"backend/src/internal/schemas"
	utils2 "backend/src/utils"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type UserService struct {
	UserRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		UserRepo: userRepo,
	}
}

var cSvc = lServices.NewCryptService()

func (s *UserService) CreateUser(user *models.RequestUser) (*string, error) {

	var hashPassword string

	hashPassword = cSvc.HashPassword(user.Password)
	newUserID := utils2.GenerateRandomID()

	dbUser := &schemas.User{
		UserID:   newUserID,
		Username: user.Username,
		Name:     user.Name,
		Email:    user.Email,
		Password: hashPassword,
		Phone:    user.Phone,
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

	return &newUserID, nil
}

func (s *UserService) GetUserByID(id string) (*models.ResponseUser, error) {
	user, err := s.UserRepo.GetUser(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	userResponse := models.ResponseUser{
		UserID:    user.UserID,
		Username:  user.Username,
		Name:      user.Name,
		Email:     user.Email,
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	return &userResponse, nil
}
