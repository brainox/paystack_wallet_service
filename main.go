package main

import (
	"log"

	"github.com/brainox/paystack_wallet_service/internal/config"
	"github.com/brainox/paystack_wallet_service/pkg/handlers"
	"github.com/brainox/paystack_wallet_service/pkg/router"
	"github.com/brainox/paystack_wallet_service/services/auth"
	"github.com/brainox/paystack_wallet_service/services/database"
	"github.com/brainox/paystack_wallet_service/services/paystack"
	"github.com/brainox/paystack_wallet_service/services/repository"
	"github.com/brainox/paystack_wallet_service/services/wallet"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	if err := database.Initialize(&cfg.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	log.Println("Database connected successfully")

	// Initialize repositories
	userRepo := repository.NewUserRepository(database.DB)
	walletRepo := repository.NewWalletRepository(database.DB)
	transactionRepo := repository.NewTransactionRepository(database.DB)
	apiKeyRepo := repository.NewAPIKeyRepository(database.DB)

	// Initialize services
	jwtService := auth.NewJWTService(cfg.JWT.Secret, cfg.JWT.ExpiryDuration)
	apiKeyService := auth.NewAPIKeyService(apiKeyRepo)
	paystackService := paystack.NewPaystackService(cfg.Paystack.SecretKey)

	googleAuthService := auth.NewGoogleAuthService(
		&cfg.Google,
		userRepo,
		walletRepo,
		jwtService,
	)

	walletService := wallet.NewWalletService(
		database.DB,
		walletRepo,
		transactionRepo,
		userRepo,
		paystackService,
	)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(googleAuthService)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyService)
	walletHandler := handlers.NewWalletHandler(walletService, paystackService)

	// Setup router
	walletRouter := router.NewWalletRouter(
		authHandler,
		apiKeyHandler,
		walletHandler,
		jwtService,
		apiKeyService,
	)

	r := walletRouter.Setup()

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Starting server on %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
