package lServices

import (
	"backend/src/internal/models"
	"backend/src/internal/repository/user"
	"backend/src/utils"
	"errors"
	"fmt"
)

var CryptSvc = utils.NewCryptService()
var AuthSvc = utils.NewAuthService()

type LoginService struct {
	UserRepo *user.UserRepository
}

func NewLoginService(userRepo *user.UserRepository) *LoginService {
	return &LoginService{
		UserRepo: userRepo,
	}
}

func (l *LoginService) Try(credentials *models.RequestLogin) (*string, error) {

	userResult, err := l.UserRepo.GetUserByUsername(credentials.Username)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}
	if userResult == nil {
		return nil, errors.New("user not found")
	}
	if !CryptSvc.ComparePassword(userResult.Password, credentials.Password) {
		return nil, errors.New("invalid password")
	}

	newToken, err := AuthSvc.GenerateJwt(userResult.UserID)
	if err != nil {
		return nil, fmt.Errorf("error generating token: %w", err)
	}

	return &newToken, err
}
