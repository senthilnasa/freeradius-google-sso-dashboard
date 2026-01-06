package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/session"
)

type contextKey string

const (
	SessionContextKey contextKey = "session"
	ClaimsContextKey  contextKey = "claims"
)

// AuthMiddleware validates JWT tokens
type AuthMiddleware struct {
	sessionService *session.Service
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(sessionService *session.Service) *AuthMiddleware {
	return &AuthMiddleware{
		sessionService: sessionService,
	}
}

// Authenticate validates the JWT token and loads session
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from cookie or Authorization header
		token := m.extractToken(r)
		if token == "" {
			log.Debug().Msg("No authentication token found")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := m.sessionService.ValidateToken(token)
		if err != nil {
			log.Debug().Err(err).Msg("Token validation failed")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get session from database
		sess, err := m.sessionService.Get(claims.SessionID)
		if err != nil {
			log.Debug().Err(err).Str("session_id", claims.SessionID).Msg("Session not found")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Update last activity
		if err := m.sessionService.UpdateActivity(sess.SessionID); err != nil {
			log.Warn().Err(err).Msg("Failed to update session activity")
		}

		// Add session and claims to context
		ctx := context.WithValue(r.Context(), SessionContextKey, sess)
		ctx = context.WithValue(ctx, ClaimsContextKey, claims)

		// Continue to next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractToken extracts JWT token from request
func (m *AuthMiddleware) extractToken(r *http.Request) string {
	// Try cookie first
	cookie, err := r.Cookie("auth_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Try Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	return ""
}

// RequireAuth is a convenience wrapper for Authenticate
func (m *AuthMiddleware) RequireAuth(handler http.HandlerFunc) http.Handler {
	return m.Authenticate(http.HandlerFunc(handler))
}
