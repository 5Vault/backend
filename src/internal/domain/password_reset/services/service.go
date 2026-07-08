package passwordResetServices

import (
	"backend/src/external"
	"backend/src/internal/logger"
	"backend/src/internal/schemas"
	passwordResetRepo "backend/src/internal/repository/password_reset"
	usrRepo "backend/src/internal/repository/user"
	"backend/src/pkg/apperr"
	"backend/src/utils"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

type PasswordResetService struct {
	Repo     *passwordResetRepo.PasswordResetRepository
	UserRepo *usrRepo.UserRepository
	Email    *external.EmailClient
}

func New(repo *passwordResetRepo.PasswordResetRepository, userRepo *usrRepo.UserRepository, email *external.EmailClient) *PasswordResetService {
	return &PasswordResetService{Repo: repo, UserRepo: userRepo, Email: email}
}

func (s *PasswordResetService) RequestReset(email string) error {
	user, err := s.UserRepo.GetUserByEmail(email)
	if err != nil || user == nil {
		// não revela se email existe ou não
		return nil
	}
	if user.AuthProvider != "local" {
		return apperr.BadRequest("conta vinculada a " + user.AuthProvider + "; use o provedor para acessar")
	}

	// invalida tokens anteriores
	_ = s.Repo.InvalidateByUserID(user.UserID)

	tokenBytes := make([]byte, 32)
	_, _ = rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	exp := time.Now().Add(1 * time.Hour)
	_ = s.Repo.Create(&schemas.PasswordResetToken{
		TokenID:   utils.GenerateULID(),
		UserID:    user.UserID,
		Token:     token,
		ExpiresAt: exp,
	})

	appURL := os.Getenv("APP_URL")
	resetURL := fmt.Sprintf("%s/reset-password/%s", appURL, token)

	go func() {
		if err := s.Email.RenderAndSend(user.Email, "Redefinição de senha — FiveKeepr", "password_reset", map[string]any{
			"ResetURL": resetURL,
		}); err != nil {
			logger.Warn("failed to send password reset email", zap.String("user_id", user.UserID), zap.Error(err))
		}
	}()

	return nil
}

func (s *PasswordResetService) Reset(token, newPassword string) error {
	t, err := s.Repo.GetValid(token)
	if err != nil {
		return apperr.BadRequest("token inválido ou expirado")
	}

	user, err := s.UserRepo.GetUserByID(t.UserID)
	if err != nil {
		return apperr.NotFound("usuário não encontrado")
	}

	cSvc := utils.NewCryptService()
	user.Password = cSvc.HashPassword(newPassword)
	if err := s.UserRepo.UpdateUser(user); err != nil {
		return apperr.Internal("erro ao atualizar senha", err)
	}

	_ = s.Repo.MarkUsed(t.TokenID)
	return nil
}
