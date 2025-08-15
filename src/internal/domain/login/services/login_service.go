package lServices

import (
	"backend/src/internal/domain/user/repositories"
	"backend/src/internal/models"
	"errors"
	"fmt"
)

var CryptSvc *CryptService = NewCryptService()
var AuthSvc *AuthService = NewAuthService()

type LoginService struct {
	UserRepo *repositories.UserRepository
}

func NewLoginService(userRepo *repositories.UserRepository) *LoginService {
	return &LoginService{
		UserRepo: userRepo,
	}
}

func (l *LoginService) Try(credentials *models.RequestLogin) (*string, error) {

	user, err := l.UserRepo.GetUserByUsername(credentials.Username)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if !CryptSvc.ComparePassword(user.Password, credentials.Password) {
		return nil, errors.New("invalid password")
	}

	newToken, err := AuthSvc.GenerateJwt(user.UserID)
	if err != nil {
		return nil, fmt.Errorf("error generating token: %w", err)
	}

	return &newToken, err
}
