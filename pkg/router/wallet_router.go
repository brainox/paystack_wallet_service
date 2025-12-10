package router

import (
	"github.com/brainox/paystack_wallet_service/internal/models"
	"github.com/brainox/paystack_wallet_service/pkg/handlers"
	"github.com/brainox/paystack_wallet_service/pkg/middleware"
	"github.com/brainox/paystack_wallet_service/services/auth"
	"github.com/gin-gonic/gin"
)

type WalletRouter struct {
	authHandler   *handlers.AuthHandler
	apiKeyHandler *handlers.APIKeyHandler
	walletHandler *handlers.WalletHandler
	jwtService    *auth.JWTService
	apiKeyService *auth.APIKeyService
}

func NewWalletRouter(
	authHandler *handlers.AuthHandler,
	apiKeyHandler *handlers.APIKeyHandler,
	walletHandler *handlers.WalletHandler,
	jwtService *auth.JWTService,
	apiKeyService *auth.APIKeyService,
) *WalletRouter {
	return &WalletRouter{
		authHandler:   authHandler,
		apiKeyHandler: apiKeyHandler,
		walletHandler: walletHandler,
		jwtService:    jwtService,
		apiKeyService: apiKeyService,
	}
}

func (r *WalletRouter) Setup() *gin.Engine {
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes (no authentication required)
	auth := router.Group("/auth")
	{
		auth.GET("/google", r.authHandler.HandleGoogleLogin)
		auth.GET("/google/callback", r.authHandler.HandleGoogleCallback)
	}

	// Webhook route (no authentication required but signature validation)
	router.POST("/wallet/paystack/webhook", r.walletHandler.HandlePaystackWebhook)

	// Authenticated routes
	authMiddleware := middleware.AuthMiddleware(r.jwtService, r.apiKeyService)

	// API Key management routes (JWT only)
	keys := router.Group("/keys")
	keys.Use(authMiddleware)
	{
		keys.POST("/create", r.apiKeyHandler.CreateAPIKey)
		keys.POST("/rollover", r.apiKeyHandler.RolloverAPIKey)
	}

	// Wallet routes
	wallet := router.Group("/wallet")
	wallet.Use(authMiddleware)
	{
		// Deposit (requires JWT or API key with deposit permission)
		wallet.POST("/deposit",
			middleware.RequirePermission(models.PermissionDeposit),
			r.walletHandler.InitiateDeposit,
		)

		// Get deposit status (read permission)
		wallet.GET("/deposit/:reference/status",
			middleware.RequirePermission(models.PermissionRead),
			r.walletHandler.GetDepositStatus,
		)

		// Get balance (read permission)
		wallet.GET("/balance",
			middleware.RequirePermission(models.PermissionRead),
			r.walletHandler.GetBalance,
		)

		// Transfer (transfer permission)
		wallet.POST("/transfer",
			middleware.RequirePermission(models.PermissionTransfer),
			r.walletHandler.Transfer,
		)

		// Transaction history (read permission)
		wallet.GET("/transactions",
			middleware.RequirePermission(models.PermissionRead),
			r.walletHandler.GetTransactionHistory,
		)
	}

	return router
}
