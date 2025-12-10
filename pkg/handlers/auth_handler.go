package handlers

import (
	"net/http"

	"github.com/brainox/paystack_wallet_service/services/auth"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	googleAuthService *auth.GoogleAuthService
}

func NewAuthHandler(googleAuthService *auth.GoogleAuthService) *AuthHandler {
	return &AuthHandler{
		googleAuthService: googleAuthService,
	}
}

// HandleGoogleLogin initiates Google OAuth login
func (h *AuthHandler) HandleGoogleLogin(c *gin.Context) {
	state := "random_state_string" // In production, use a secure random state
	url := h.googleAuthService.GetAuthURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// HandleGoogleCallback handles the OAuth callback from Google
func (h *AuthHandler) HandleGoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code not found"})
		return
	}

	token, err := h.googleAuthService.HandleCallback(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}
