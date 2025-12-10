package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/brainox/paystack_wallet_service/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TransactionRepository struct {
	db *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(tx *sqlx.Tx, transaction *models.Transaction) error {
	query := `
		INSERT INTO transactions (
			id, user_id, wallet_id, type, amount, status, reference, 
			paystack_reference, recipient_wallet_id, recipient_user_id, 
			description, metadata, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at
	`
	transaction.ID = uuid.New()
	transaction.CreatedAt = time.Now()
	transaction.UpdatedAt = time.Now()

	return tx.QueryRow(
		query,
		transaction.ID,
		transaction.UserID,
		transaction.WalletID,
		transaction.Type,
		transaction.Amount,
		transaction.Status,
		transaction.Reference,
		transaction.PaystackReference,
		transaction.RecipientWalletID,
		transaction.RecipientUserID,
		transaction.Description,
		transaction.Metadata,
		transaction.CreatedAt,
		transaction.UpdatedAt,
	).Scan(&transaction.ID, &transaction.CreatedAt, &transaction.UpdatedAt)
}

func (r *TransactionRepository) GetByID(id uuid.UUID) (*models.Transaction, error) {
	var transaction models.Transaction
	query := `SELECT * FROM transactions WHERE id = $1`
	err := r.db.Get(&transaction, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, err
	}
	return &transaction, nil
}

func (r *TransactionRepository) GetByReference(reference string) (*models.Transaction, error) {
	var transaction models.Transaction
	query := `SELECT * FROM transactions WHERE reference = $1`
	err := r.db.Get(&transaction, query, reference)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, err
	}
	return &transaction, nil
}

func (r *TransactionRepository) GetByPaystackReference(paystackReference string) (*models.Transaction, error) {
	var transaction models.Transaction
	query := `SELECT * FROM transactions WHERE paystack_reference = $1`
	err := r.db.Get(&transaction, query, paystackReference)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, err
	}
	return &transaction, nil
}

func (r *TransactionRepository) GetByUserID(userID uuid.UUID, limit, offset int) ([]models.Transaction, error) {
	var transactions []models.Transaction
	query := `
		SELECT * FROM transactions 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`
	err := r.db.Select(&transactions, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *TransactionRepository) UpdateStatus(tx *sqlx.Tx, id uuid.UUID, status models.TransactionStatus) error {
	query := `
		UPDATE transactions
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := tx.Exec(query, status, time.Now(), id)
	return err
}

func (r *TransactionRepository) UpdateStatusByReference(reference string, status models.TransactionStatus) error {
	query := `
		UPDATE transactions
		SET status = $1, updated_at = $2
		WHERE reference = $3
	`
	_, err := r.db.Exec(query, status, time.Now(), reference)
	return err
}
