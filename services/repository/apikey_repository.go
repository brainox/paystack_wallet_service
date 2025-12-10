package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/brainox/paystack_wallet_service/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type APIKeyRepository struct {
	db *sqlx.DB
}

func NewAPIKeyRepository(db *sqlx.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(apiKey *models.APIKey) error {
	query := `
		INSERT INTO api_keys (
			id, user_id, name, key_hash, key_prefix, permissions, 
			expires_at, is_active, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	apiKey.ID = uuid.New()
	apiKey.CreatedAt = time.Now()
	apiKey.UpdatedAt = time.Now()

	return r.db.QueryRow(
		query,
		apiKey.ID,
		apiKey.UserID,
		apiKey.Name,
		apiKey.KeyHash,
		apiKey.KeyPrefix,
		apiKey.Permissions,
		apiKey.ExpiresAt,
		apiKey.IsActive,
		apiKey.CreatedAt,
		apiKey.UpdatedAt,
	).Scan(&apiKey.ID, &apiKey.CreatedAt, &apiKey.UpdatedAt)
}

func (r *APIKeyRepository) GetByID(id uuid.UUID) (*models.APIKey, error) {
	var apiKey models.APIKey
	query := `SELECT * FROM api_keys WHERE id = $1`
	err := r.db.Get(&apiKey, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("API key not found")
		}
		return nil, err
	}
	return &apiKey, nil
}

func (r *APIKeyRepository) GetByKeyHash(keyHash string) (*models.APIKey, error) {
	var apiKey models.APIKey
	query := `SELECT * FROM api_keys WHERE key_hash = $1`
	err := r.db.Get(&apiKey, query, keyHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("API key not found")
		}
		return nil, err
	}
	return &apiKey, nil
}

func (r *APIKeyRepository) GetByUserID(userID uuid.UUID) ([]models.APIKey, error) {
	var apiKeys []models.APIKey
	query := `SELECT * FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC`
	err := r.db.Select(&apiKeys, query, userID)
	if err != nil {
		return nil, err
	}
	return apiKeys, nil
}

func (r *APIKeyRepository) CountActiveKeys(userID uuid.UUID) (int, error) {
	var count int
	query := `SELECT count_active_api_keys($1)`
	err := r.db.Get(&count, query, userID)
	return count, err
}

func (r *APIKeyRepository) Revoke(id uuid.UUID) error {
	query := `
		UPDATE api_keys
		SET is_active = false, revoked_at = $1, updated_at = $2
		WHERE id = $3
	`
	now := time.Now()
	_, err := r.db.Exec(query, now, now, id)
	return err
}

func (r *APIKeyRepository) UpdateLastUsed(id uuid.UUID) error {
	query := `
		UPDATE api_keys
		SET last_used_at = $1, updated_at = $2
		WHERE id = $3
	`
	now := time.Now()
	_, err := r.db.Exec(query, now, now, id)
	return err
}

func (r *APIKeyRepository) GetExpiredKeyByID(id uuid.UUID) (*models.APIKey, error) {
	var apiKey models.APIKey
	query := `SELECT * FROM api_keys WHERE id = $1 AND expires_at < NOW()`
	err := r.db.Get(&apiKey, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("expired API key not found")
		}
		return nil, err
	}
	return &apiKey, nil
}
