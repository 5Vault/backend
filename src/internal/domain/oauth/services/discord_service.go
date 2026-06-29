package oauthServices

import (
	"backend/src/internal/logger"
	usrRepo "backend/src/internal/repository/user"
	"backend/src/pkg/apperr"
	"backend/src/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

var discordEndpoint = oauth2.Endpoint{
	AuthURL:  "https://discord.com/api/oauth2/authorize",
	TokenURL: "https://discord.com/api/oauth2/token",
}

type DiscordUserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
}

func (u *DiscordUserInfo) AvatarURL() string {
	if u.Avatar == "" {
		return ""
	}
	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png?size=256", u.ID, u.Avatar)
}

type DiscordUserCreator interface {
	CreateDiscordUser(email, name, discordID, avatarURL string) (*string, error)
}

type DiscordOAuthService struct {
	UserRepo *usrRepo.UserRepository
	UserSvc  DiscordUserCreator
	AuthSvc  *utils.AuthService
	Config   *oauth2.Config
}

func NewDiscordOAuthService(userRepo *usrRepo.UserRepository, userSvc DiscordUserCreator) *DiscordOAuthService {
	return &DiscordOAuthService{
		UserRepo: userRepo,
		UserSvc:  userSvc,
		AuthSvc:  utils.NewAuthService(),
		Config: &oauth2.Config{
			ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
			ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("DISCORD_REDIRECT_URL"),
			Scopes:       []string{"identify", "email"},
			Endpoint:     discordEndpoint,
		},
	}
}

func (s *DiscordOAuthService) GetAuthURL(state string) string {
	return s.Config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (s *DiscordOAuthService) HandleCallback(ctx context.Context, code string) (string, error) {
	token, err := s.Config.Exchange(ctx, code)
	if err != nil {
		logger.Error("discord token exchange failed", zap.Error(err))
		return "", apperr.Internal("authentication failed", err)
	}

	resp, err := newHTTPClientWithToken(token.AccessToken).Get("https://discord.com/api/users/@me")
	if err != nil {
		return "", apperr.Internal("failed to retrieve discord user info", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", apperr.Internal("failed to read discord user info", err)
	}

	var du DiscordUserInfo
	if err := json.Unmarshal(body, &du); err != nil {
		return "", apperr.Internal("failed to parse discord user info", err)
	}

	if du.Email == "" {
		return "", apperr.BadRequest("conta Discord sem email verificado")
	}

	userID, err := s.findOrCreate(du)
	if err != nil {
		return "", err
	}

	jwt, err := s.AuthSvc.GenerateJwt(userID)
	if err != nil {
		return "", apperr.Internal("failed to generate token", err)
	}

	logger.Info("discord oauth login", zap.String("user_id", userID), zap.String("discord_id", du.ID))
	return jwt, nil
}

func (s *DiscordOAuthService) findOrCreate(du DiscordUserInfo) (string, error) {
	existing, err := s.UserRepo.GetUserByDiscordID(du.ID)
	if err == nil && existing != nil {
		return existing.UserID, nil
	}

	existing, err = s.UserRepo.GetUserByEmail(du.Email)
	if err == nil && existing != nil {
		did := du.ID
		existing.DiscordID = &did
		if existing.AuthProvider == "local" {
			existing.AuthProvider = "discord"
		}
		av := du.AvatarURL()
		if av != "" && existing.AvatarURL == nil {
			existing.AvatarURL = &av
		}
		_ = s.UserRepo.UpdateUser(existing)
		return existing.UserID, nil
	}

	userID, err := s.UserSvc.CreateDiscordUser(du.Email, du.Username, du.ID, du.AvatarURL())
	if err != nil {
		return "", err
	}
	return *userID, nil
}
