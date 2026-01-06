package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/radius"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/session"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/pkg/logger"
)

type APIHandler struct {
	sessionService *session.Service
	radiusClient   *radius.Client
	db             *sqlx.DB
	logger         *logger.Logger
}

func NewAPIHandler(
	sessionService *session.Service,
	radiusClient *radius.Client,
	db *sqlx.DB,
	logger *logger.Logger,
) *APIHandler {
	return &APIHandler{
		sessionService: sessionService,
		radiusClient:   radiusClient,
		db:             db,
		logger:         logger,
	}
}

func (h *APIHandler) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Implementation for RADIUS authorization callback
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "authorized",
	})
}

func (h *APIHandler) HandleAccounting(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID string `json:"session_id"`
		Type      string `json:"type"` // start, update, stop
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Implementation for RADIUS accounting callback
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "recorded",
	})
}

func (h *APIHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	sess := r.Context().Value("session")
	if sess == nil {
		http.Error(w, "No session found", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sess)
}
