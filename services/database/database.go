package database

import (
	"fmt"

	"github.com/brainox/paystack_wallet_service/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// DB holds the database connection
var DB *sqlx.DB

// Initialize initializes the database connection
func Initialize(cfg *config.DatabaseConfig) error {
	var err error
	DB, err = sqlx.Connect("postgres", cfg.GetDSN())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
