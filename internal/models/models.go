package models

import (
	"database/sql/driver"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	GoogleID  *string   `json:"google_id,omitempty" db:"google_id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Wallet represents a user's wallet
type Wallet struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	WalletNumber string    `json:"wallet_number" db:"wallet_number"`
	Balance      float64   `json:"balance" db:"balance"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeDeposit  TransactionType = "deposit"
	TransactionTypeTransfer TransactionType = "transfer"
	TransactionTypeCredit   TransactionType = "credit"
	TransactionTypeDebit    TransactionType = "debit"
)

func (t *TransactionType) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		*t = TransactionType(string(v))
	case string:
		*t = TransactionType(v)
	}
	return nil
}

func (t TransactionType) Value() (driver.Value, error) {
	return string(t), nil
}

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending TransactionStatus = "pending"
	TransactionStatusSuccess TransactionStatus = "success"
	TransactionStatusFailed  TransactionStatus = "failed"
)

func (s *TransactionStatus) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		*s = TransactionStatus(string(v))
	case string:
		*s = TransactionStatus(v)
	}
	return nil
}

func (s TransactionStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// Transaction represents a wallet transaction
type Transaction struct {
	ID                uuid.UUID         `json:"id" db:"id"`
	UserID            uuid.UUID         `json:"user_id" db:"user_id"`
	WalletID          uuid.UUID         `json:"wallet_id" db:"wallet_id"`
	Type              TransactionType   `json:"type" db:"type"`
	Amount            float64           `json:"amount" db:"amount"`
	Status            TransactionStatus `json:"status" db:"status"`
	Reference         *string           `json:"reference,omitempty" db:"reference"`
	PaystackReference *string           `json:"paystack_reference,omitempty" db:"paystack_reference"`
	RecipientWalletID *uuid.UUID        `json:"recipient_wallet_id,omitempty" db:"recipient_wallet_id"`
	RecipientUserID   *uuid.UUID        `json:"recipient_user_id,omitempty" db:"recipient_user_id"`
	Description       *string           `json:"description,omitempty" db:"description"`
	Metadata          *string           `json:"metadata,omitempty" db:"metadata"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at" db:"updated_at"`
}

// APIKey represents an API key for service-to-service access
type APIKey struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	UserID      uuid.UUID      `json:"user_id" db:"user_id"`
	Name        string         `json:"name" db:"name"`
	KeyHash     string         `json:"-" db:"key_hash"`
	KeyPrefix   string         `json:"key_prefix" db:"key_prefix"`
	Permissions pq.StringArray `json:"permissions" db:"permissions"`
	ExpiresAt   time.Time      `json:"expires_at" db:"expires_at"`
	IsActive    bool           `json:"is_active" db:"is_active"`
	RevokedAt   *time.Time     `json:"revoked_at,omitempty" db:"revoked_at"`
	LastUsedAt  *time.Time     `json:"last_used_at,omitempty" db:"last_used_at"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

// Valid permissions for API keys
const (
	PermissionDeposit  = "deposit"
	PermissionTransfer = "transfer"
	PermissionRead     = "read"
)

// IsValidPermission checks if a permission is valid
func IsValidPermission(permission string) bool {
	validPermissions := map[string]bool{
		PermissionDeposit:  true,
		PermissionTransfer: true,
		PermissionRead:     true,
	}
	return validPermissions[permission]
}

// HasPermission checks if the API key has a specific permission
func (a *APIKey) HasPermission(permission string) bool {
	for _, p := range a.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// IsExpired checks if the API key is expired
func (a *APIKey) IsExpired() bool {
	return time.Now().After(a.ExpiresAt)
}

// IsRevoked checks if the API key is revoked
func (a *APIKey) IsRevoked() bool {
	return a.RevokedAt != nil
}

// IsValid checks if the API key is valid (not expired, not revoked, and active)
func (a *APIKey) IsValid() bool {
	return a.IsActive && !a.IsExpired() && !a.IsRevoked()
}
