package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/brainox/paystack_wallet_service/external/external_models"
	"github.com/brainox/paystack_wallet_service/internal/models"
	"github.com/brainox/paystack_wallet_service/pkg/middleware"
	"github.com/brainox/paystack_wallet_service/services/paystack"
	"github.com/brainox/paystack_wallet_service/services/wallet"
	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	walletService   *wallet.WalletService
	paystackService *paystack.PaystackService
}

func NewWalletHandler(
	walletService *wallet.WalletService,
	paystackService *paystack.PaystackService,
) *WalletHandler {
	return &WalletHandler{
		walletService:   walletService,
		paystackService: paystackService,
	}
}

type DepositRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type DepositResponse struct {
	Reference        string `json:"reference"`
	AuthorizationURL string `json:"authorization_url"`
}

// InitiateDeposit initiates a wallet deposit
func (h *WalletHandler) InitiateDeposit(c *gin.Context) {
	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	reference, authURL, err := h.walletService.InitiateDeposit(userID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DepositResponse{
		Reference:        reference,
		AuthorizationURL: authURL,
	})
}

// HandlePaystackWebhook handles Paystack webhook events
func (h *WalletHandler) HandlePaystackWebhook(c *gin.Context) {
	// Get the signature from headers
	signature := c.GetHeader("x-paystack-signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing signature"})
		return
	}

	// Read the raw body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Validate the signature
	if !h.paystackService.ValidateWebhookSignature(body, signature) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Parse the webhook event
	var event external_models.WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook payload"})
		return
	}

	// Process the webhook
	if err := h.walletService.ProcessWebhook(&event); err != nil {
		// Log error but return 200 to prevent Paystack from retrying
		c.JSON(http.StatusOK, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true})
}

// GetDepositStatus gets the status of a deposit
func (h *WalletHandler) GetDepositStatus(c *gin.Context) {
	reference := c.Param("reference")
	if reference == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reference is required"})
		return
	}

	transaction, err := h.walletService.GetDepositStatus(reference)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reference": transaction.Reference,
		"status":    transaction.Status,
		"amount":    transaction.Amount,
	})
}

// GetBalance gets the wallet balance
func (h *WalletHandler) GetBalance(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	balance, err := h.walletService.GetBalance(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"balance": balance})
}

// GetWalletInfo gets the wallet information including wallet number
func (h *WalletHandler) GetWalletInfo(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	wallet, err := h.walletService.GetWalletDetails(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wallet_number": wallet.WalletNumber,
		"balance":       wallet.Balance,
		"created_at":    wallet.CreatedAt,
	})
}

type TransferRequest struct {
	WalletNumber string  `json:"wallet_number" binding:"required"`
	Amount       float64 `json:"amount" binding:"required,gt=0"`
}

// Transfer transfers money to another wallet
func (h *WalletHandler) Transfer(c *gin.Context) {
	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := h.walletService.Transfer(userID, req.WalletNumber, req.Amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Transfer completed",
	})
}

type TransactionHistoryResponse struct {
	Type   models.TransactionType   `json:"type"`
	Amount float64                  `json:"amount"`
	Status models.TransactionStatus `json:"status"`
}

// GetTransactionHistory gets the transaction history
func (h *WalletHandler) GetTransactionHistory(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get pagination parameters
	limit := 50
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	transactions, err := h.walletService.GetTransactionHistory(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format response
	response := make([]TransactionHistoryResponse, len(transactions))
	for i, txn := range transactions {
		response[i] = TransactionHistoryResponse{
			Type:   txn.Type,
			Amount: txn.Amount,
			Status: txn.Status,
		}
	}

	c.JSON(http.StatusOK, response)
}
