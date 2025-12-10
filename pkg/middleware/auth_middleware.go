package middleware

import (
	"net/http"
	"strings"

	"github.com/brainox/paystack_wallet_service/services/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

const (
	UserIDKey            = "user_id"
	UserEmailKey         = "user_email"
	APIKeyPermissionsKey = "api_key_permissions"
	IsAPIKeyAuth         = "is_api_key_auth"
)

// AuthMiddleware handles both JWT and API key authentication
func AuthMiddleware(jwtService *auth.JWTService, apiKeyService *auth.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for API key first
		apiKey := c.GetHeader("x-api-key")
		if apiKey != "" {
			// Validate API key
			apiKeyModel, err := apiKeyService.ValidateAPIKey(apiKey)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired API key"})
				c.Abort()
				return
			}

			// Set user context
			c.Set(UserIDKey, apiKeyModel.UserID)
			c.Set(APIKeyPermissionsKey, apiKeyModel.Permissions)
			c.Set(IsAPIKeyAuth, true)
			c.Next()
			return
		}

		// Check for JWT token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate JWT token
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user context
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)
		c.Set(IsAPIKeyAuth, false)
		c.Next()
	}
}

// RequirePermission checks if the API key has the required permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		isAPIKey, exists := c.Get(IsAPIKeyAuth)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication context not found"})
			c.Abort()
			return
		}

		// If JWT auth, allow all permissions
		if !isAPIKey.(bool) {
			c.Next()
			return
		}

		// Check API key permissions
		permissions, exists := c.Get(APIKeyPermissionsKey)
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "No permissions found"})
			c.Abort()
			return
		}

		hasPermission := false
		// Handle both pq.StringArray and []string types
		switch perms := permissions.(type) {
		case pq.StringArray:
			for _, perm := range perms {
				if perm == permission {
					hasPermission = true
					break
				}
			}
		case []string:
			for _, perm := range perms {
				if perm == permission {
					hasPermission = true
					break
				}
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID extracts the user ID from the context
func GetUserID(c *gin.Context) (uuid.UUID, error) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return uuid.Nil, gin.Error{Err: http.ErrAbortHandler, Type: gin.ErrorTypePrivate}
	}
	return userID.(uuid.UUID), nil
}
