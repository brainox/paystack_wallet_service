package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/brainox/paystack_wallet_service/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type WalletRepository struct {
	db *sqlx.DB
}

func NewWalletRepository(db *sqlx.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) Create(wallet *models.Wallet) error {
	query := `
		INSERT INTO wallets (id, user_id, wallet_number, balance, created_at, updated_at)
		VALUES ($1, $2, generate_wallet_number(), $3, $4, $5)
		RETURNING id, wallet_number, created_at, updated_at
	`
	wallet.ID = uuid.New()
	wallet.CreatedAt = time.Now()
	wallet.UpdatedAt = time.Now()

	return r.db.QueryRow(
		query,
		wallet.ID,
		wallet.UserID,
		wallet.Balance,
		wallet.CreatedAt,
		wallet.UpdatedAt,
	).Scan(&wallet.ID, &wallet.WalletNumber, &wallet.CreatedAt, &wallet.UpdatedAt)
}

func (r *WalletRepository) GetByID(id uuid.UUID) (*models.Wallet, error) {
	var wallet models.Wallet
	query := `SELECT * FROM wallets WHERE id = $1`
	err := r.db.Get(&wallet, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wallet not found")
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *WalletRepository) GetByUserID(userID uuid.UUID) (*models.Wallet, error) {
	var wallet models.Wallet
	query := `SELECT * FROM wallets WHERE user_id = $1`
	err := r.db.Get(&wallet, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wallet not found")
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *WalletRepository) GetByWalletNumber(walletNumber string) (*models.Wallet, error) {
	var wallet models.Wallet
	query := `SELECT * FROM wallets WHERE wallet_number = $1`
	err := r.db.Get(&wallet, query, walletNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wallet not found")
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *WalletRepository) UpdateBalance(tx *sqlx.Tx, walletID uuid.UUID, newBalance float64) error {
	query := `
		UPDATE wallets
		SET balance = $1, updated_at = $2
		WHERE id = $3 AND balance >= 0
	`
	result, err := tx.Exec(query, newBalance, time.Now(), walletID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("insufficient balance or wallet not found")
	}

	return nil
}

func (r *WalletRepository) GetBalanceForUpdate(tx *sqlx.Tx, walletID uuid.UUID) (float64, error) {
	var balance float64
	query := `SELECT balance FROM wallets WHERE id = $1 FOR UPDATE`
	err := tx.Get(&balance, query, walletID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("wallet not found")
		}
		return 0, err
	}
	return balance, nil
}
