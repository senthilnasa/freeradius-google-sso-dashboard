package database

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/config"
)

// Connect establishes a connection to the database
func Connect(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
	dsn := cfg.DSN()

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.Name).
		Msg("Connecting to database")

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to database")
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.MaxConns)
	db.SetMaxIdleConns(cfg.MaxIdle)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := db.Ping(); err != nil {
		log.Error().Err(err).Msg("Failed to ping database")
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	log.Info().Msg("Database connection established successfully")

	return db, nil
}

// Close closes the database connection
func Close(db *sqlx.DB) error {
	if db != nil {
		log.Info().Msg("Closing database connection")
		return db.Close()
	}
	return nil
}

// HealthCheck checks if the database is accessible
func HealthCheck(db *sqlx.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}
