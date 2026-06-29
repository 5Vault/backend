package lServices

import (
	"backend/src/external"
	"backend/src/internal/actionlog"
	"backend/src/internal/logger"
	"backend/src/internal/models"
	actionLogRepo "backend/src/internal/repository/action_log"
	"backend/src/internal/repository/user"
	"backend/src/pkg/apperr"
	"backend/src/utils"
	"fmt"
	"os"
	"time"

	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
)

func totpValidate(code, secret string) bool {
	return totp.Validate(code, secret)
}

var CryptSvc = utils.NewCryptService()
var AuthSvc = utils.NewAuthService()

type LoginService struct {
	UserRepo      *user.UserRepository
	ActionLogRepo *actionLogRepo.ActionLogRepository
	Email         *external.EmailClient
}

func NewLoginService(userRepo *user.UserRepository, alRepo *actionLogRepo.ActionLogRepository, email *external.EmailClient) *LoginService {
	return &LoginService{UserRepo: userRepo, ActionLogRepo: alRepo, Email: email}
}

func (l *LoginService) Try(credentials *models.RequestLogin, ip string) (*string, error) {
	userResult, err := l.UserRepo.GetUserByUsername(credentials.Username)
	if err != nil || userResult == nil {
		return nil, apperr.Unauthorized("invalid credentials")
	}

	if !CryptSvc.ComparePassword(userResult.Password, credentials.Password) {
		logger.Warn("failed login attempt", zap.String("username", credentials.Username))
		actionlog.Log(userResult.UserID, "login.failed", "", "", ip)
		return nil, apperr.Unauthorized("invalid credentials")
	}

	if userResult.TwoFAEnabled && userResult.TwoFASecret != nil {
		if credentials.TwoFACode == "" {
			return nil, apperr.NewAppError(499, "2fa_required")
		}
		if !totpValidate(credentials.TwoFACode, *userResult.TwoFASecret) {
			return nil, apperr.Unauthorized("código 2FA inválido")
		}
	}

	token, err := AuthSvc.GenerateJwt(userResult.UserID)
	if err != nil {
		logger.Error("failed to generate token", zap.String("user_id", userResult.UserID), zap.Error(err))
		return nil, apperr.Internal("error generating token", err)
	}

	go func() {
		isNewDevice := !l.ActionLogRepo.HasLoginFromIP(userResult.UserID, ip)
		actionlog.Log(userResult.UserID, "login", "", "", ip)
		if isNewDevice && ip != "" && ip != "::1" && ip != "127.0.0.1" {
			actionlog.Log(userResult.UserID, "login.new_device", "", "", ip)
			appURL := os.Getenv("APP_URL")
			if err := l.Email.RenderAndSend(userResult.Email, "Novo acesso à sua conta FiveVault", "new_device", map[string]any{
				"IP":       ip,
				"Time":     time.Now().Format("02/01/2006 às 15:04"),
				"ResetURL": fmt.Sprintf("%s/forgot-password", appURL),
			}); err != nil {
				logger.Warn("failed to send new_device email", zap.String("user_id", userResult.UserID), zap.Error(err))
			}
		}
	}()

	logger.Info("user logged in", zap.String("user_id", userResult.UserID), zap.String("ip", ip))
	return &token, nil
}
