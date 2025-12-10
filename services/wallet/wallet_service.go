package wallet

import (
	"fmt"
	"time"

	"github.com/brainox/paystack_wallet_service/external/external_models"
	"github.com/brainox/paystack_wallet_service/internal/models"
	"github.com/brainox/paystack_wallet_service/services/paystack"
	"github.com/brainox/paystack_wallet_service/services/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type WalletService struct {
	db              *sqlx.DB
	walletRepo      *repository.WalletRepository
	transactionRepo *repository.TransactionRepository
	userRepo        *repository.UserRepository
	paystackService *paystack.PaystackService
}

func NewWalletService(
	db *sqlx.DB,
	walletRepo *repository.WalletRepository,
	transactionRepo *repository.TransactionRepository,
	userRepo *repository.UserRepository,
	paystackService *paystack.PaystackService,
) *WalletService {
	return &WalletService{
		db:              db,
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
		paystackService: paystackService,
	}
}

// InitiateDeposit initiates a deposit using Paystack
func (s *WalletService) InitiateDeposit(userID uuid.UUID, amount float64) (string, string, error) {
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get user: %w", err)
	}

	// Get user's wallet
	wallet, err := s.walletRepo.GetByUserID(userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get wallet: %w", err)
	}

	// Generate unique reference
	reference := fmt.Sprintf("DEP_%s_%d", uuid.New().String()[:8], time.Now().Unix())

	// Convert amount to kobo (Paystack uses smallest currency unit)
	amountInKobo := int(amount * 100)

	// Initialize Paystack transaction
	paystackResp, err := s.paystackService.InitializeTransaction(user.Email, amountInKobo, reference)
	if err != nil {
		return "", "", fmt.Errorf("failed to initialize Paystack transaction: %w", err)
	}

	// Create pending transaction record
	tx, err := s.db.Beginx()
	if err != nil {
		return "", "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	transaction := &models.Transaction{
		UserID:            userID,
		WalletID:          wallet.ID,
		Type:              models.TransactionTypeDeposit,
		Amount:            amount,
		Status:            models.TransactionStatusPending,
		Reference:         &reference,
		PaystackReference: &paystackResp.Data.Reference,
		Description:       stringPtr("Wallet deposit via Paystack"),
	}

	if err := s.transactionRepo.Create(tx, transaction); err != nil {
		return "", "", fmt.Errorf("failed to create transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return reference, paystackResp.Data.AuthorizationURL, nil
}

// ProcessWebhook processes a Paystack webhook event
func (s *WalletService) ProcessWebhook(event *external_models.WebhookEvent) error {
	// Only process successful charge events
	if event.Event != "charge.success" {
		return nil
	}

	// Get transaction by Paystack reference
	transaction, err := s.transactionRepo.GetByPaystackReference(event.Data.Reference)
	if err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	// Check if already processed (idempotency)
	if transaction.Status == models.TransactionStatusSuccess {
		return nil // Already processed
	}

	// Verify the transaction status from Paystack
	verifyResp, err := s.paystackService.VerifyTransaction(event.Data.Reference)
	if err != nil {
		return fmt.Errorf("failed to verify transaction: %w", err)
	}

	if verifyResp.Data.Status != "success" {
		return fmt.Errorf("transaction not successful")
	}

	// Begin database transaction
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get wallet with lock
	balance, err := s.walletRepo.GetBalanceForUpdate(tx, transaction.WalletID)
	if err != nil {
		return fmt.Errorf("failed to get wallet balance: %w", err)
	}

	// Calculate new balance
	newBalance := balance + transaction.Amount

	// Update wallet balance
	if err := s.walletRepo.UpdateBalance(tx, transaction.WalletID, newBalance); err != nil {
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}

	// Update transaction status
	if err := s.transactionRepo.UpdateStatus(tx, transaction.ID, models.TransactionStatusSuccess); err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetDepositStatus gets the status of a deposit transaction
func (s *WalletService) GetDepositStatus(reference string) (*models.Transaction, error) {
	return s.transactionRepo.GetByReference(reference)
}

// GetBalance gets the wallet balance for a user
func (s *WalletService) GetBalance(userID uuid.UUID) (float64, error) {
	wallet, err := s.walletRepo.GetByUserID(userID)
	if err != nil {
		return 0, err
	}
	return wallet.Balance, nil
}

// Transfer transfers money from one wallet to another
func (s *WalletService) Transfer(senderUserID uuid.UUID, recipientWalletNumber string, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}

	// Get sender's wallet
	senderWallet, err := s.walletRepo.GetByUserID(senderUserID)
	if err != nil {
		return fmt.Errorf("failed to get sender wallet: %w", err)
	}

	// Get recipient's wallet
	recipientWallet, err := s.walletRepo.GetByWalletNumber(recipientWalletNumber)
	if err != nil {
		return fmt.Errorf("recipient wallet not found: %w", err)
	}

	// Check if sender is trying to send to themselves
	if senderWallet.ID == recipientWallet.ID {
		return fmt.Errorf("cannot transfer to your own wallet")
	}

	// Begin database transaction
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get sender balance with lock
	senderBalance, err := s.walletRepo.GetBalanceForUpdate(tx, senderWallet.ID)
	if err != nil {
		return fmt.Errorf("failed to get sender balance: %w", err)
	}

	// Check sufficient balance
	if senderBalance < amount {
		return fmt.Errorf("insufficient balance")
	}

	// Get recipient balance with lock
	recipientBalance, err := s.walletRepo.GetBalanceForUpdate(tx, recipientWallet.ID)
	if err != nil {
		return fmt.Errorf("failed to get recipient balance: %w", err)
	}

	// Update sender balance (debit)
	newSenderBalance := senderBalance - amount
	if err := s.walletRepo.UpdateBalance(tx, senderWallet.ID, newSenderBalance); err != nil {
		return fmt.Errorf("failed to update sender balance: %w", err)
	}

	// Update recipient balance (credit)
	newRecipientBalance := recipientBalance + amount
	if err := s.walletRepo.UpdateBalance(tx, recipientWallet.ID, newRecipientBalance); err != nil {
		return fmt.Errorf("failed to update recipient balance: %w", err)
	}

	// Create debit transaction for sender
	baseReference := fmt.Sprintf("TXF_%s_%d", uuid.New().String()[:8], time.Now().Unix())
	debitReference := fmt.Sprintf("%s_DEBIT", baseReference)
	debitTransaction := &models.Transaction{
		UserID:            senderUserID,
		WalletID:          senderWallet.ID,
		Type:              models.TransactionTypeDebit,
		Amount:            amount,
		Status:            models.TransactionStatusSuccess,
		Reference:         &debitReference,
		RecipientWalletID: &recipientWallet.ID,
		RecipientUserID:   &recipientWallet.UserID,
		Description:       stringPtr(fmt.Sprintf("Transfer to wallet %s", recipientWalletNumber)),
	}
	if err := s.transactionRepo.Create(tx, debitTransaction); err != nil {
		return fmt.Errorf("failed to create debit transaction: %w", err)
	}

	// Create credit transaction for recipient
	creditReference := fmt.Sprintf("%s_CREDIT", baseReference)
	creditTransaction := &models.Transaction{
		UserID:      recipientWallet.UserID,
		WalletID:    recipientWallet.ID,
		Type:        models.TransactionTypeCredit,
		Amount:      amount,
		Status:      models.TransactionStatusSuccess,
		Reference:   &creditReference,
		Description: stringPtr(fmt.Sprintf("Transfer from wallet %s", senderWallet.WalletNumber)),
	}
	if err := s.transactionRepo.Create(tx, creditTransaction); err != nil {
		return fmt.Errorf("failed to create credit transaction: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetTransactionHistory gets the transaction history for a user
func (s *WalletService) GetTransactionHistory(userID uuid.UUID, limit, offset int) ([]models.Transaction, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	return s.transactionRepo.GetByUserID(userID, limit, offset)
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
