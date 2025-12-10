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

type APIKeyInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	KeyPrefix   string   `json:"key_prefix"`
	Permissions []string `json:"permissions"`
	ExpiresAt   string   `json:"expires_at"`
	IsActive    bool     `json:"is_active"`
	CreatedAt   string   `json:"created_at"`
}

// ListAPIKeys lists all API keys for the user
func (h *APIKeyHandler) ListAPIKeys(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	apiKeys, err := h.apiKeyService.ListAPIKeys(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []APIKeyInfo
	for _, key := range apiKeys {
		response = append(response, APIKeyInfo{
			ID:          key.ID.String(),
			Name:        key.Name,
			KeyPrefix:   key.KeyPrefix,
			Permissions: key.Permissions,
			ExpiresAt:   key.ExpiresAt.Format("2006-01-02T15:04:05Z"),
			IsActive:    key.IsActive,
			CreatedAt:   key.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, response)
}

// DeleteAPIKey deletes (revokes) an API key
func (h *APIKeyHandler) DeleteAPIKey(c *gin.Context) {
	keyID := c.Param("id")
	parsedKeyID, err := uuid.Parse(keyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API key ID"})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	err = h.apiKeyService.DeleteAPIKey(userID, parsedKeyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key deleted successfully"})
}
