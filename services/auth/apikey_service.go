package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/brainox/paystack_wallet_service/internal/models"
	"github.com/brainox/paystack_wallet_service/services/repository"
	"github.com/google/uuid"
)

const (
	MaxActiveAPIKeys = 5
	APIKeyPrefix     = "sk_live_"
	APIKeyLength     = 32
)

type APIKeyService struct {
	repo *repository.APIKeyRepository
}

func NewAPIKeyService(repo *repository.APIKeyRepository) *APIKeyService {
	return &APIKeyService{repo: repo}
}

// ParseExpiry converts expiry string (1H, 1D, 1M, 1Y) to duration
func ParseExpiry(expiryStr string) (time.Duration, error) {
	if len(expiryStr) < 2 {
		return 0, fmt.Errorf("invalid expiry format")
	}

	unit := expiryStr[len(expiryStr)-1:]
	value := expiryStr[:len(expiryStr)-1]

	var duration time.Duration
	switch strings.ToUpper(unit) {
	case "H":
		duration = time.Hour
	case "D":
		duration = 24 * time.Hour
	case "M":
		duration = 30 * 24 * time.Hour
	case "Y":
		duration = 365 * 24 * time.Hour
	default:
		return 0, fmt.Errorf("invalid expiry unit: must be H, D, M, or Y")
	}

	// Parse the numeric value
	var multiplier int
	_, err := fmt.Sscanf(value, "%d", &multiplier)
	if err != nil || multiplier <= 0 {
		return 0, fmt.Errorf("invalid expiry value")
	}

	return time.Duration(multiplier) * duration, nil
}

// GenerateAPIKey generates a secure random API key
func (s *APIKeyService) GenerateAPIKey() (string, error) {
	bytes := make([]byte, APIKeyLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return APIKeyPrefix + hex.EncodeToString(bytes), nil
}

// HashAPIKey creates a SHA-256 hash of the API key
func (s *APIKeyService) HashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

// CreateAPIKey creates a new API key for a user
func (s *APIKeyService) CreateAPIKey(userID uuid.UUID, name string, permissions []string, expiryStr string) (string, *models.APIKey, error) {
	// Validate permissions
	for _, perm := range permissions {
		if !models.IsValidPermission(perm) {
			return "", nil, fmt.Errorf("invalid permission: %s", perm)
		}
	}

	// Check active key count
	activeCount, err := s.repo.CountActiveKeys(userID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to count active keys: %w", err)
	}

	if activeCount >= MaxActiveAPIKeys {
		return "", nil, fmt.Errorf("maximum of %d active API keys allowed", MaxActiveAPIKeys)
	}

	// Parse expiry
	duration, err := ParseExpiry(expiryStr)
	if err != nil {
		return "", nil, err
	}

	// Generate API key
	apiKey, err := s.GenerateAPIKey()
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Hash the key for storage
	keyHash := s.HashAPIKey(apiKey)

	// Create API key record
	apiKeyModel := &models.APIKey{
		UserID:      userID,
		Name:        name,
		KeyHash:     keyHash,
		KeyPrefix:   APIKeyPrefix,
		Permissions: permissions,
		ExpiresAt:   time.Now().Add(duration),
		IsActive:    true,
	}

	if err := s.repo.Create(apiKeyModel); err != nil {
		return "", nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return apiKey, apiKeyModel, nil
}

func (s *APIKeyService) ListAPIKeys(userID uuid.UUID) ([]models.APIKey, error) {
	return s.repo.GetByUserID(userID)
}

func (s *APIKeyService) DeleteAPIKey(userID uuid.UUID, keyID uuid.UUID) error {
	return s.RevokeAPIKey(keyID, userID)
}

// ValidateAPIKey validates an API key and returns the associated key record
func (s *APIKeyService) ValidateAPIKey(apiKey string) (*models.APIKey, error) {
	keyHash := s.HashAPIKey(apiKey)
	apiKeyModel, err := s.repo.GetByKeyHash(keyHash)
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	if !apiKeyModel.IsValid() {
		return nil, fmt.Errorf("API key is expired, revoked, or inactive")
	}

	// Update last used timestamp
	_ = s.repo.UpdateLastUsed(apiKeyModel.ID)

	return apiKeyModel, nil
}

// RolloverAPIKey creates a new API key using the same permissions as an expired key
func (s *APIKeyService) RolloverAPIKey(userID uuid.UUID, expiredKeyID uuid.UUID, expiryStr string) (string, *models.APIKey, error) {
	// Get the expired key
	expiredKey, err := s.repo.GetExpiredKeyByID(expiredKeyID)
	if err != nil {
		return "", nil, fmt.Errorf("expired key not found or not expired")
	}

	// Verify the key belongs to the user
	if expiredKey.UserID != userID {
		return "", nil, fmt.Errorf("unauthorized: key does not belong to user")
	}

	// Check active key count
	activeCount, err := s.repo.CountActiveKeys(userID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to count active keys: %w", err)
	}

	if activeCount >= MaxActiveAPIKeys {
		return "", nil, fmt.Errorf("maximum of %d active API keys allowed", MaxActiveAPIKeys)
	}

	// Parse expiry
	duration, err := ParseExpiry(expiryStr)
	if err != nil {
		return "", nil, err
	}

	// Generate new API key with same permissions
	apiKey, err := s.GenerateAPIKey()
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	keyHash := s.HashAPIKey(apiKey)

	// Create new API key with same permissions
	newAPIKey := &models.APIKey{
		UserID:      userID,
		Name:        expiredKey.Name + " (rolled over)",
		KeyHash:     keyHash,
		KeyPrefix:   APIKeyPrefix,
		Permissions: expiredKey.Permissions,
		ExpiresAt:   time.Now().Add(duration),
		IsActive:    true,
	}

	if err := s.repo.Create(newAPIKey); err != nil {
		return "", nil, fmt.Errorf("failed to create rolled over API key: %w", err)
	}

	return apiKey, newAPIKey, nil
}

// RevokeAPIKey revokes an API key
func (s *APIKeyService) RevokeAPIKey(keyID uuid.UUID, userID uuid.UUID) error {
	apiKey, err := s.repo.GetByID(keyID)
	if err != nil {
		return err
	}

	if apiKey.UserID != userID {
		return fmt.Errorf("unauthorized: key does not belong to user")
	}

	return s.repo.Revoke(keyID)
}
