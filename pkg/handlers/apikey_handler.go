package handlers

import (
	"net/http"

	"github.com/brainox/paystack_wallet_service/pkg/middleware"
	"github.com/brainox/paystack_wallet_service/services/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type APIKeyHandler struct {
	apiKeyService *auth.APIKeyService
}

func NewAPIKeyHandler(apiKeyService *auth.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService: apiKeyService,
	}
}

type CreateAPIKeyRequest struct {
	Name        string   `json:"name" binding:"required"`
	Permissions []string `json:"permissions" binding:"required"`
	Expiry      string   `json:"expiry" binding:"required"`
}

type CreateAPIKeyResponse struct {
	APIKey    string `json:"api_key"`
	ExpiresAt string `json:"expires_at"`
}

// CreateAPIKey creates a new API key
func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	apiKey, apiKeyModel, err := h.apiKeyService.CreateAPIKey(userID, req.Name, req.Permissions, req.Expiry)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, CreateAPIKeyResponse{
		APIKey:    apiKey,
		ExpiresAt: apiKeyModel.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	})
}

type RolloverAPIKeyRequest struct {
	ExpiredKeyID string `json:"expired_key_id" binding:"required"`
	Expiry       string `json:"expiry" binding:"required"`
}

// RolloverAPIKey creates a new API key from an expired one
func (h *APIKeyHandler) RolloverAPIKey(c *gin.Context) {
	var req RolloverAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	expiredKeyID, err := uuid.Parse(req.ExpiredKeyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expired_key_id"})
		return
	}

	apiKey, apiKeyModel, err := h.apiKeyService.RolloverAPIKey(userID, expiredKeyID, req.Expiry)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, CreateAPIKeyResponse{
		APIKey:    apiKey,
		ExpiresAt: apiKeyModel.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	})
}
