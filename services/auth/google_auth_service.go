package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brainox/paystack_wallet_service/internal/config"
	"github.com/brainox/paystack_wallet_service/internal/models"
	"github.com/brainox/paystack_wallet_service/services/repository"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleAuthService struct {
	config     *oauth2.Config
	userRepo   *repository.UserRepository
	walletRepo *repository.WalletRepository
	jwtService *JWTService
}

func NewGoogleAuthService(
	cfg *config.GoogleOAuthConfig,
	userRepo *repository.UserRepository,
	walletRepo *repository.WalletRepository,
	jwtService *JWTService,
) *GoogleAuthService {
	return &GoogleAuthService{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
		userRepo:   userRepo,
		walletRepo: walletRepo,
		jwtService: jwtService,
	}
}

func (s *GoogleAuthService) GetAuthURL(state string) string {
	return s.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (s *GoogleAuthService) HandleCallback(ctx context.Context, code string) (string, error) {
	token, err := s.config.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code: %w", err)
	}

	client := s.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", fmt.Errorf("failed to decode user info: %w", err)
	}

	// Get or create user
	user, err := s.userRepo.GetByGoogleID(userInfo.ID)
	if err != nil {
		// User doesn't exist, create new one
		user = &models.User{
			Email:    userInfo.Email,
			GoogleID: &userInfo.ID,
			Name:     userInfo.Name,
		}
		if err := s.userRepo.Create(user); err != nil {
			return "", fmt.Errorf("failed to create user: %w", err)
		}

		// Create wallet for new user
		wallet := &models.Wallet{
			UserID:  user.ID,
			Balance: 0,
		}
		if err := s.walletRepo.Create(wallet); err != nil {
			return "", fmt.Errorf("failed to create wallet: %w", err)
		}
	}

	// Generate JWT token
	jwtToken, err := s.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	return jwtToken, nil
}
