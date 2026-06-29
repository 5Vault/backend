package oauthServices

import (
	"backend/src/internal/domain/user/services"
	"backend/src/internal/logger"
	usrRepo "backend/src/internal/repository/user"
	"backend/src/pkg/apperr"
	"backend/src/utils"
	"context"
	"encoding/json"
	"io"
	"os"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type OAuthService struct {
	UserRepo    *usrRepo.UserRepository
	UserService *services.UserService
	AuthSvc     *utils.AuthService
	Config      *oauth2.Config
}

func NewOAuthService(userRepo *usrRepo.UserRepository, userService *services.UserService) *OAuthService {
	config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	return &OAuthService{
		UserRepo:    userRepo,
		UserService: userService,
		AuthSvc:     utils.NewAuthService(),
		Config:      config,
	}
}

func (s *OAuthService) GetGoogleAuthURL(state string) string {
	return s.Config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (s *OAuthService) HandleGoogleCallback(ctx context.Context, code string) (string, error) {
	token, err := s.Config.Exchange(ctx, code)
	if err != nil {
		logger.Error("google token exchange failed", zap.Error(err))
		return "", apperr.Internal("authentication failed", err)
	}

	client := s.Config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		logger.Error("failed to fetch google userinfo", zap.Error(err))
		return "", apperr.Internal("failed to retrieve user info", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", apperr.Internal("failed to read user info", err)
	}

	var googleUser GoogleUserInfo
	if err := json.Unmarshal(body, &googleUser); err != nil {
		return "", apperr.Internal("failed to parse user info", err)
	}

	if googleUser.Email == "" {
		return "", apperr.BadRequest("google account email not available")
	}

	userID, err := s.findOrCreateGoogleUser(googleUser)
	if err != nil {
		return "", err
	}

	jwt, err := s.AuthSvc.GenerateJwt(userID)
	if err != nil {
		logger.Error("failed to generate jwt", zap.String("user_id", userID), zap.Error(err))
		return "", apperr.Internal("failed to generate token", err)
	}

	logger.Info("google oauth login", zap.String("user_id", userID), zap.String("email", googleUser.Email))
	return jwt, nil
}

func (s *OAuthService) findOrCreateGoogleUser(info GoogleUserInfo) (string, error) {
	existing, err := s.UserRepo.GetUserByGoogleID(info.ID)
	if err == nil && existing != nil {
		return existing.UserID, nil
	}

	existing, err = s.UserRepo.GetUserByEmail(info.Email)
	if err == nil && existing != nil {
		gid := info.ID
		existing.GoogleID = &gid
		if existing.AuthProvider == "local" {
			existing.AuthProvider = "google"
		}
		_ = s.UserRepo.UpdateUser(existing)
		return existing.UserID, nil
	}

	userID, err := s.UserService.CreateGoogleUser(info.Email, info.Name, info.ID)
	if err != nil {
		return "", err
	}
	return *userID, nil
}
