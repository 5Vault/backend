package services

import (
	"backend/src/external"
	"backend/src/internal/actionlog"
	"backend/src/internal/logger"
	keyRepo "backend/src/internal/repository/key"
	usrRepo "backend/src/internal/repository/user"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"backend/src/internal/models"
	"backend/src/internal/schemas"
	"backend/src/pkg/apperr"
	"backend/src/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"

	tierService "backend/src/internal/domain/user/tier/service"
)

var tierSvc = tierService.NewTierService()
var cSvc = utils.NewCryptService()

type UserService struct {
	UserRepo      *usrRepo.UserRepository
	KeyRepo       *keyRepo.KeyRepository
	Email         *external.EmailClient
	OnUserCreated func(ctx context.Context, userID string)
}

func NewUserService(userRepo *usrRepo.UserRepository, keyRepo *keyRepo.KeyRepository, email *external.EmailClient) *UserService {
	return &UserService{UserRepo: userRepo, KeyRepo: keyRepo, Email: email}
}

func (s *UserService) CreateUser(user *models.RequestUser) (*string, error) {
	if _, err := s.UserRepo.GetUserByUsername(user.Username); err == nil {
		return nil, apperr.Conflict("username already exists")
	}
	if _, err := s.UserRepo.GetUserByEmail(user.Email); err == nil {
		return nil, apperr.Conflict("email already exists")
	}

	var phone *string
	if user.Phone != "" {
		if _, err := s.UserRepo.GetUserByPhone(user.Phone); err == nil {
			return nil, apperr.Conflict("phone already exists")
		}
		phone = &user.Phone
	}

	newUserID := utils.GenerateULID()
	dbUser := &schemas.User{
		UserID:        newUserID,
		Username:      user.Username,
		Name:          user.Name,
		Email:         user.Email,
		Password:      cSvc.HashPassword(user.Password),
		Phone:         phone,
		Tier:          "free",
		TierUpdatedAt: time.Now(),
		AuthProvider:  "local",
	}

	if err := s.UserRepo.CreateUser(dbUser); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint") {
			return nil, apperr.Conflict("user already exists")
		}
		logger.Error("failed to create user", zap.String("username", user.Username), zap.Error(err))
		return nil, apperr.Internal("failed to create user", err)
	}

	if err := s.createAPIKey(newUserID); err != nil {
		logger.Error("failed to create api key", zap.String("user_id", newUserID), zap.Error(err))
		return nil, apperr.Internal("failed to create api key", err)
	}

	actionlog.Log(newUserID, "user.register", "user", newUserID, "", nil)

	if s.OnUserCreated != nil {
		go s.OnUserCreated(context.Background(), newUserID)
	}

	go func() {
		appURL := os.Getenv("APP_URL")
		if err := s.Email.RenderAndSend(user.Email, "Bem-vindo ao FiveVault!", "welcome", map[string]any{
			"Name":     user.Name,
			"Username": user.Username,
			"AppURL":   fmt.Sprintf("%s/dashboard", appURL),
		}); err != nil {
			logger.Warn("failed to send welcome email", zap.String("user_id", newUserID), zap.Error(err))
		}
	}()

	logger.Info("user created", zap.String("user_id", newUserID), zap.String("username", user.Username))
	return &newUserID, nil
}

func (s *UserService) CreateGoogleUser(email, name, googleID string) (*string, error) {
	newUserID := utils.GenerateULID()

	base := strings.Split(email, "@")[0]
	base = strings.ToLower(strings.ReplaceAll(base, ".", "_"))
	username := base
	if _, err := s.UserRepo.GetUserByUsername(username); err == nil {
		username = base + "_" + utils.GenerateULID()[:6]
	}

	gid := googleID
	dbUser := &schemas.User{
		UserID:        newUserID,
		Username:      username,
		Name:          name,
		Email:         email,
		GoogleID:      &gid,
		Tier:          "free",
		TierUpdatedAt: time.Now(),
		AuthProvider:  "google",
	}

	if err := s.UserRepo.CreateUser(dbUser); err != nil {
		logger.Error("failed to create google user", zap.String("email", email), zap.Error(err))
		return nil, apperr.Internal("failed to create user", err)
	}

	if err := s.createAPIKey(newUserID); err != nil {
		return nil, apperr.Internal("failed to create api key", err)
	}

	if s.OnUserCreated != nil {
		go s.OnUserCreated(context.Background(), newUserID)
	}
	logger.Info("google user created", zap.String("user_id", newUserID), zap.String("email", email))
	return &newUserID, nil
}

func (s *UserService) CreateDiscordUser(email, name, discordID, avatarURL string) (*string, error) {
	newUserID := utils.GenerateULID()

	base := strings.Split(email, "@")[0]
	base = strings.ToLower(strings.ReplaceAll(base, ".", "_"))
	username := base
	if _, err := s.UserRepo.GetUserByUsername(username); err == nil {
		username = base + "_" + utils.GenerateULID()[:6]
	}

	did := discordID
	var av *string
	if avatarURL != "" {
		av = &avatarURL
	}
	dbUser := &schemas.User{
		UserID:        newUserID,
		Username:      username,
		Name:          name,
		Email:         email,
		DiscordID:     &did,
		AvatarURL:     av,
		Tier:          "free",
		TierUpdatedAt: time.Now(),
		AuthProvider:  "discord",
	}

	if err := s.UserRepo.CreateUser(dbUser); err != nil {
		logger.Error("failed to create discord user", zap.String("email", email), zap.Error(err))
		return nil, apperr.Internal("failed to create user", err)
	}

	if err := s.createAPIKey(newUserID); err != nil {
		return nil, apperr.Internal("failed to create api key", err)
	}

	if s.OnUserCreated != nil {
		go s.OnUserCreated(context.Background(), newUserID)
	}
	logger.Info("discord user created", zap.String("user_id", newUserID), zap.String("email", email))
	return &newUserID, nil
}

func (s *UserService) UpdateAvatar(userID, url string) error {
	return s.UserRepo.UpdateAvatarURL(userID, url)
}

func (s *UserService) SetExtraStorage(userID string, enabled bool) error {
	user, err := s.UserRepo.GetUserByID(userID)
	if err != nil {
		return apperr.NotFound("user not found")
	}
	user.ExtraStorageEnabled = enabled
	return s.UserRepo.UpdateUser(user)
}

func (s *UserService) createAPIKey(userID string) error {
	key, err := utils.GenerateAPIKey()
	if err != nil {
		return err
	}
	return s.KeyRepo.New(&schemas.Key{Key: key, UserID: userID})
}

func (s *UserService) GetUserByID(id string, own bool) (*models.ResponseUser, error) {
	user, err := s.UserRepo.GetUser(id)
	if err != nil {
		return nil, apperr.NotFound("user not found")
	}

	fmtTime := func(t *time.Time) string {
		if t == nil {
			return ""
		}
		return t.Format("2006-01-02 15:04:05")
	}

	resp := models.ResponseUser{
		UserID:              user.UserID,
		Username:            user.Username,
		Name:                user.Name,
		Email:               user.Email,
		Tier:                user.Tier,
		TierName:            tierSvc.GetTierNameByID(user.Tier),
		TierUpdatedAt:       user.TierUpdatedAt.Format("2006-01-02 15:04:05"),
		ExtraStorageEnabled: user.ExtraStorageEnabled,
		TwoFAEnabled:        user.TwoFAEnabled,
		AvatarURL:           user.AvatarURL,
		CreatedAt:           fmtTime(user.CreatedAt),
		UpdatedAt:           fmtTime(user.UpdatedAt),
	}

	if user.Phone != nil {
		resp.Phone = *user.Phone
	}

	if own {
		key, err := s.KeyRepo.GetByUserID(id)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error("error retrieving api key", zap.String("user_id", id), zap.Error(err))
			return nil, apperr.Internal("error retrieving api key", err)
		}
		if key != nil {
			resp.ApiKey = &key.Key
		}
	}

	return &resp, nil
}

func (s *UserService) GetRawUser(userID string) (*schemas.User, error) {
	return s.UserRepo.GetUserByID(userID)
}

func (s *UserService) Set2FASecret(userID, secret string) error {
	return s.UserRepo.Set2FASecret(userID, secret)
}

func (s *UserService) Enable2FA(userID string) error {
	return s.UserRepo.Enable2FA(userID)
}

func (s *UserService) Disable2FA(userID string) error {
	return s.UserRepo.Disable2FA(userID)
}
