package session

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/config"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/models"
)

// Service handles session management
type Service struct {
	db            *sqlx.DB
	jwtSecret     []byte
	sessionTimeout time.Duration
}

// Claims represents JWT claims
type Claims struct {
	SessionID string `json:"session_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	VLANId    int    `json:"vlan_id"`
	UserType  string `json:"user_type"`
	jwt.RegisteredClaims
}

// NewService creates a new session service
func NewService(db *sqlx.DB, cfg *config.SecurityConfig) *Service {
	return &Service{
		db:             db,
		jwtSecret:      []byte(cfg.JWTSecret),
		sessionTimeout: cfg.SessionTimeout,
	}
}

// Create creates a new session
func (s *Service) Create(userInfo *models.UserInfo, vlanAssignment *models.VLANAssignment, ipAddress, macAddress, nasIP, userAgent string) (*models.Session, string, error) {
	// Generate session ID
	sessionID := generateSessionID()

	// Extract domain from email
	emailDomain := extractDomain(userInfo.Email)

	// Create session record
	session := &models.Session{
		SessionID:   sessionID,
		Username:    userInfo.Email,
		Email:       userInfo.Email,
		EmailDomain: emailDomain,
		VLANId:      vlanAssignment.VLANId,
		UserType:    vlanAssignment.UserType,
		IsActive:    true,
	}

	// Set optional fields
	if ipAddress != "" {
		session.IPAddress = sql.NullString{String: ipAddress, Valid: true}
	}
	if macAddress != "" {
		session.MACAddress = sql.NullString{String: macAddress, Valid: true}
	}
	if nasIP != "" {
		session.NASIp = sql.NullString{String: nasIP, Valid: true}
	}
	if userAgent != "" {
		session.UserAgent = sql.NullString{String: userAgent, Valid: true}
	}

	// Set expiry time
	expiresAt := time.Now().Add(s.sessionTimeout)
	session.ExpiresAt = sql.NullTime{Time: expiresAt, Valid: true}

	// Insert into database
	query := `
		INSERT INTO sessions (
			session_id, username, email, email_domain, vlan_id, user_type,
			ip_address, mac_address, nas_ip, user_agent, expires_at, is_active
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	result, err := s.db.Exec(query,
		session.SessionID,
		session.Username,
		session.Email,
		session.EmailDomain,
		session.VLANId,
		session.UserType,
		session.IPAddress,
		session.MACAddress,
		session.NASIp,
		session.UserAgent,
		session.ExpiresAt,
		session.IsActive,
	)

	if err != nil {
		log.Error().Err(err).Str("email", userInfo.Email).Msg("Failed to create session")
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get session ID")
		return nil, "", fmt.Errorf("failed to get session ID: %w", err)
	}

	session.ID = id
	session.CreatedAt = time.Now()
	session.LastActivity = time.Now()

	// Generate JWT token
	token, err := s.GenerateToken(session)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate JWT token")
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	log.Info().
		Str("session_id", sessionID).
		Str("email", userInfo.Email).
		Int("vlan_id", vlanAssignment.VLANId).
		Str("user_type", vlanAssignment.UserType).
		Msg("Session created successfully")

	return session, token, nil
}

// GenerateToken generates a JWT token for a session
func (s *Service) GenerateToken(session *models.Session) (string, error) {
	expiresAt := time.Now().Add(s.sessionTimeout)

	claims := &Claims{
		SessionID: session.SessionID,
		Email:     session.Email,
		Username:  session.Username,
		VLANId:    session.VLANId,
		UserType:  session.UserType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   session.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		log.Debug().Err(err).Msg("Token validation failed")
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// Get retrieves a session by session ID
func (s *Service) Get(sessionID string) (*models.Session, error) {
	var session models.Session
	query := `
		SELECT * FROM sessions
		WHERE session_id = ? AND is_active = 1
		LIMIT 1
	`

	err := s.db.Get(&session, query, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		log.Error().Err(err).Str("session_id", sessionID).Msg("Failed to get session")
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Check if session is expired
	if session.ExpiresAt.Valid && session.ExpiresAt.Time.Before(time.Now()) {
		log.Debug().Str("session_id", sessionID).Msg("Session expired")
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

// GetByEmail retrieves active sessions for an email
func (s *Service) GetByEmail(email string) ([]models.Session, error) {
	var sessions []models.Session
	query := `
		SELECT * FROM sessions
		WHERE email = ? AND is_active = 1
		ORDER BY created_at DESC
	`

	err := s.db.Select(&sessions, query, email)
	if err != nil {
		log.Error().Err(err).Str("email", email).Msg("Failed to get sessions by email")
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	return sessions, nil
}

// UpdateActivity updates the last activity timestamp
func (s *Service) UpdateActivity(sessionID string) error {
	query := `
		UPDATE sessions
		SET last_activity = NOW()
		WHERE session_id = ? AND is_active = 1
	`

	result, err := s.db.Exec(query, sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionID).Msg("Failed to update session activity")
		return fmt.Errorf("failed to update activity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found or inactive")
	}

	log.Debug().Str("session_id", sessionID).Msg("Session activity updated")
	return nil
}

// Terminate terminates a session
func (s *Service) Terminate(sessionID string) error {
	query := `
		UPDATE sessions
		SET is_active = 0
		WHERE session_id = ? AND is_active = 1
	`

	result, err := s.db.Exec(query, sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionID).Msg("Failed to terminate session")
		return fmt.Errorf("failed to terminate session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found or already inactive")
	}

	log.Info().Str("session_id", sessionID).Msg("Session terminated")
	return nil
}

// TerminateByEmail terminates all sessions for an email
func (s *Service) TerminateByEmail(email string) error {
	query := `
		UPDATE sessions
		SET is_active = 0
		WHERE email = ? AND is_active = 1
	`

	result, err := s.db.Exec(query, email)
	if err != nil {
		log.Error().Err(err).Str("email", email).Msg("Failed to terminate sessions by email")
		return fmt.Errorf("failed to terminate sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	log.Info().
		Str("email", email).
		Int64("sessions_terminated", rowsAffected).
		Msg("Sessions terminated by email")

	return nil
}

// CleanupExpired removes expired sessions
func (s *Service) CleanupExpired() (int64, error) {
	query := `
		UPDATE sessions
		SET is_active = 0
		WHERE is_active = 1
		AND (
			expires_at < NOW()
			OR last_activity < DATE_SUB(NOW(), INTERVAL 30 MINUTE)
		)
	`

	result, err := s.db.Exec(query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to cleanup expired sessions")
		return 0, fmt.Errorf("failed to cleanup sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected > 0 {
		log.Info().
			Int64("sessions_cleaned", rowsAffected).
			Msg("Expired sessions cleaned up")
	}

	return rowsAffected, nil
}

// GetActiveSessions returns count of active sessions
func (s *Service) GetActiveSessions() (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM sessions
		WHERE is_active = 1
		AND expires_at > NOW()
	`

	err := s.db.Get(&count, query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get active sessions count")
		return 0, fmt.Errorf("failed to get active sessions: %w", err)
	}

	return count, nil
}

// Helper functions

func generateSessionID() string {
	// Generate a unique session ID
	// Using timestamp + random string for simplicity
	// In production, use crypto/rand for better randomness
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("sess_%d_%s", timestamp, randomString(16))
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(1 * time.Nanosecond) // Simple way to get variation
	}
	return string(b)
}

func extractDomain(email string) string {
	parts := []byte(email)
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == '@' {
			return string(parts[i+1:])
		}
	}
	return ""
}
