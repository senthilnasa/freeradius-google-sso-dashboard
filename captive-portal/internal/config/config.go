package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Database configuration
	Database DatabaseConfig

	// Google OAuth configuration
	Google GoogleOAuthConfig

	// RADIUS configuration
	Radius RadiusConfig

	// Security configuration
	Security SecurityConfig

	// Application configuration
	App AppConfig

	// Branding configuration
	Branding BrandingConfig
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	MaxConns int
	MaxIdle  int
}

// GoogleOAuthConfig holds Google OAuth settings
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	HostedDomain string // Optional: Restrict to specific domain
	Scopes       []string
}

// RadiusConfig holds RADIUS server settings
type RadiusConfig struct {
	Server   string
	Secret   string
	AuthPort int
	AcctPort int
	Timeout  time.Duration
	Retries  int
}

// SecurityConfig holds security-related settings
type SecurityConfig struct {
	JWTSecret     string
	CSRFSecret    string
	SessionSecret string
	SessionTimeout time.Duration
	SecureCookie  bool
	HTTPOnly      bool
	SameSite      string
}

// AppConfig holds application settings
type AppConfig struct {
	Env      string
	Port     string
	URL      string
	LogLevel string
	LogFile  string
}

// BrandingConfig holds branding customization
type BrandingConfig struct {
	LogoURL        string
	InstitutionName string
	PrimaryColor   string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file (ignore error if not present)
	_ = godotenv.Load()

	cfg := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "mysql"),
			Port:     getEnvAsInt("DB_PORT", 3306),
			Name:     getEnv("DB_NAME", "radius"),
			User:     getEnv("DB_USER", "radius"),
			Password: getEnv("DB_PASSWORD", "radiuspassword"),
			MaxConns: getEnvAsInt("DB_MAX_CONNS", 25),
			MaxIdle:  getEnvAsInt("DB_MAX_IDLE", 5),
		},

		Google: GoogleOAuthConfig{
			ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURI:  getEnv("GOOGLE_REDIRECT_URI", ""),
			HostedDomain: getEnv("GOOGLE_HOSTED_DOMAIN", ""),
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
		},

		Radius: RadiusConfig{
			Server:   getEnv("RADIUS_SERVER", "freeradius"),
			Secret:   getEnv("RADIUS_SECRET", "testing123"),
			AuthPort: getEnvAsInt("RADIUS_AUTH_PORT", 1812),
			AcctPort: getEnvAsInt("RADIUS_ACCT_PORT", 1813),
			Timeout:  time.Duration(getEnvAsInt("RADIUS_TIMEOUT", 5)) * time.Second,
			Retries:  getEnvAsInt("RADIUS_RETRIES", 3),
		},

		Security: SecurityConfig{
			JWTSecret:      getEnv("JWT_SECRET", ""),
			CSRFSecret:     getEnv("CSRF_SECRET", ""),
			SessionSecret:  getEnv("SESSION_SECRET", ""),
			SessionTimeout: time.Duration(getEnvAsInt("SESSION_TIMEOUT", 28800)) * time.Second,
			SecureCookie:   getEnv("APP_ENV", "production") == "production",
			HTTPOnly:       true,
			SameSite:       getEnv("COOKIE_SAMESITE", "Lax"),
		},

		App: AppConfig{
			Env:      getEnv("APP_ENV", "production"),
			Port:     getEnv("APP_PORT", "8080"),
			URL:      getEnv("APP_URL", "http://localhost:8080"),
			LogLevel: getEnv("LOG_LEVEL", "info"),
			LogFile:  getEnv("LOG_FILE", "/app/logs/captive-portal.log"),
		},

		Branding: BrandingConfig{
			LogoURL:        getEnv("BRANDING_LOGO_URL", "https://via.placeholder.com/200x60?text=Institution+Logo"),
			InstitutionName: getEnv("BRANDING_INSTITUTION_NAME", "Network Authentication"),
			PrimaryColor:   getEnv("BRANDING_PRIMARY_COLOR", "#667eea"),
		},
	}

	// Validate required configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if all required configuration is present
func (c *Config) Validate() error {
	// Check Google OAuth
	if c.Google.ClientID == "" {
		return fmt.Errorf("GOOGLE_CLIENT_ID is required")
	}
	if c.Google.ClientSecret == "" {
		return fmt.Errorf("GOOGLE_CLIENT_SECRET is required")
	}
	if c.Google.RedirectURI == "" {
		return fmt.Errorf("GOOGLE_REDIRECT_URI is required")
	}

	// Check Security secrets
	if c.Security.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.Security.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	if c.Security.CSRFSecret == "" {
		return fmt.Errorf("CSRF_SECRET is required")
	}

	if c.Security.SessionSecret == "" {
		return fmt.Errorf("SESSION_SECRET is required")
	}

	// Check Database
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}

	// Check RADIUS
	if c.Radius.Secret == "" {
		return fmt.Errorf("RADIUS_SECRET is required")
	}

	return nil
}

// DSN returns the database connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
	)
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool retrieves an environment variable as a boolean or returns a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// IsDevelopment returns true if running in development mode
func (c *AppConfig) IsDevelopment() bool {
	return c.Env == "development" || c.Env == "dev"
}

// IsProduction returns true if running in production mode
func (c *AppConfig) IsProduction() bool {
	return c.Env == "production" || c.Env == "prod"
}
