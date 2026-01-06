package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/auth"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/models"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/radius"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/session"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/vlan"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/pkg/logger"
)

type AuthHandler struct {
	googleAuth     *auth.GoogleAuth
	vlanService    *vlan.Service
	sessionService *session.Service
	radiusClient   *radius.Client
	logger         *logger.Logger
}

func NewAuthHandler(
	googleAuth *auth.GoogleAuth,
	vlanService *vlan.Service,
	sessionService *session.Service,
	radiusClient *radius.Client,
	logger *logger.Logger,
) *AuthHandler {
	return &AuthHandler{
		googleAuth:     googleAuth,
		vlanService:    vlanService,
		sessionService: sessionService,
		radiusClient:   radiusClient,
		logger:         logger,
	}
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate random state for CSRF protection
	state := "random-state-token" // In production, use crypto/rand
	authURL := h.googleAuth.GetAuthURL(state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		h.logger.Error("No authorization code in callback")
		http.Error(w, "No authorization code", http.StatusBadRequest)
		return
	}

	// Exchange code for token and get user info
	userInfo, err := h.googleAuth.AuthenticateUser(context.Background(), code)
	if err != nil {
		h.logger.Error("Failed to authenticate user", "error", err)
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	h.logger.Info("User authenticated", "email", userInfo.Email)

	// Assign VLAN
	vlanAssignment, err := h.vlanService.AssignVLAN(userInfo.Email)
	if err != nil {
		h.logger.Error("Failed to assign VLAN", "email", userInfo.Email, "error", err)
		http.Error(w, "VLAN assignment failed", http.StatusInternalServerError)
		return
	}

	h.logger.Info("VLAN assigned", "email", userInfo.Email, "vlan", vlanAssignment.VLANId, "type", vlanAssignment.UserType)

	// Get client IP
	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = forwarded
	}

	// Get MAC address (if available from header)
	macAddress := r.Header.Get("X-Client-MAC")
	if macAddress == "" {
		macAddress = "00:00:00:00:00:00" // Default if not available
	}

	// NAS IP (could be from config or header)
	nasIP := "127.0.0.1" // Default, should come from configuration

	// Create session
	sess, token, err := h.sessionService.Create(userInfo, vlanAssignment, clientIP, macAddress, nasIP, r.UserAgent())
	if err != nil {
		h.logger.Error("Failed to create session", "error", err)
		http.Error(w, "Session creation failed", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Session created", "sessionID", sess.SessionID, "email", userInfo.Email)

	// Send RADIUS authentication
	err = h.radiusClient.Authenticate(context.Background(), userInfo.Email, vlanAssignment.VLANId, vlanAssignment.UserType)
	if err != nil {
		h.logger.Error("RADIUS authentication failed", "error", err)
		// Continue anyway - session is created
	}

	// Send RADIUS accounting start
	err = h.radiusClient.StartAccounting(context.Background(), sess.SessionID, userInfo.Email, clientIP)
	if err != nil {
		h.logger.Error("RADIUS accounting start failed", "error", err)
	}

	// Set session cookie
	// Note: Secure flag disabled for local HTTP testing
	// TODO: Re-enable Secure: true in production with HTTPS
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token", // Must match middleware/auth.go:75
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Disabled for HTTP testing
		SameSite: http.SameSiteLaxMode, // Changed from Strict to Lax for better compatibility
		MaxAge:   28800, // 8 hours
	})

	http.Redirect(w, r, "/success", http.StatusSeeOther)
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Get session from context
	sess := r.Context().Value("session")
	if sess == nil {
		// No session, clear cookie anyway and redirect
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			MaxAge:   -1,
			SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	sessionData, ok := sess.(*models.Session)
	if !ok {
		// Invalid session type, clear cookie and redirect
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			MaxAge:   -1,
			SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	h.logger.Info("User logging out", "email", sessionData.Email, "session_id", sessionData.SessionID)

	// Terminate session in database FIRST
	err := h.sessionService.Terminate(sessionData.SessionID)
	if err != nil {
		h.logger.Error("Failed to terminate session", "error", err)
	}

	// Send RADIUS accounting stop
	// Calculate session time
	sessionTime := int64(time.Since(sessionData.CreatedAt).Seconds())

	err = h.radiusClient.StopAccounting(
		context.Background(),
		sessionData.Email,
		sessionData.SessionID,
		sessionData.IPAddress.String,
		sessionTime,
		0, // inputOctets - would need to track this
		0, // outputOctets - would need to track this
	)
	if err != nil {
		h.logger.Error("RADIUS accounting stop failed", "error", err)
	}

	// Clear cookie - MUST match the cookie set in HandleCallback
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	h.logger.Info("Logout successful", "email", sessionData.Email)

	// Redirect to landing page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
