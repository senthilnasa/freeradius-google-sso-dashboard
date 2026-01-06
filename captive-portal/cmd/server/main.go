package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/auth"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/config"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/handlers"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/middleware"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/radius"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/session"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/vlan"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/pkg/database"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/pkg/logger"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger := logger.New(cfg.App.LogLevel)
	appLogger.Info("Starting FreeRADIUS Google SSO Captive Portal...")

	// Connect to database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		appLogger.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()
	appLogger.Info("Connected to database successfully")

	// Initialize services
	googleAuth := auth.NewGoogleAuth(&cfg.Google)
	vlanService := vlan.NewService(db)
	sessionService := session.NewService(db, &cfg.Security)
	radiusClient := radius.NewClient(&cfg.Radius)

	// Start background workers (if the method exists)
	// go sessionService.StartCleanupWorker(context.Background(), 5*time.Minute)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(
		googleAuth,
		vlanService,
		sessionService,
		radiusClient,
		appLogger,
	)
	portalHandler := handlers.NewPortalHandler(
		sessionService,
		appLogger,
		cfg.Branding.LogoURL,
		cfg.Branding.InstitutionName,
		cfg.Branding.PrimaryColor,
	)
	apiHandler := handlers.NewAPIHandler(sessionService, radiusClient, db, appLogger)

	// Setup router
	router := mux.NewRouter()

	// CSRF protection - exempt GET requests from CSRF (logout is GET)
	csrfMiddleware := csrf.Protect(
		[]byte(cfg.Security.CSRFSecret),
		csrf.Secure(cfg.App.Env == "production"),
		csrf.Path("/"),
		csrf.RequestHeader("X-CSRF-Token"),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log CSRF failures for debugging
			appLogger.Warn("CSRF validation failed", "path", r.URL.Path, "method", r.Method)
			http.Error(w, "Forbidden - CSRF validation failed", http.StatusForbidden)
		})),
	)

	// Public routes (no authentication required)
	router.HandleFunc("/", portalHandler.LandingPage).Methods("GET")
	router.HandleFunc("/login", authHandler.HandleLogin).Methods("GET")
	router.HandleFunc("/callback", authHandler.HandleCallback).Methods("GET")
	router.HandleFunc("/health", portalHandler.HealthCheck).Methods("GET")
	router.HandleFunc("/success", portalHandler.SuccessPage).Methods("GET") // Temporarily public for testing

	// Create auth middleware
	authMware := middleware.NewAuthMiddleware(sessionService)

	// Logout route - requires auth but no CSRF (will be added before CSRF middleware)
	logoutRouter := mux.NewRouter()
	logoutRouter.Use(func(next http.Handler) http.Handler {
		return authMware.Authenticate(next)
	})
	logoutRouter.HandleFunc("/logout", authHandler.HandleLogout).Methods("GET")

	// Protected routes (authentication required)
	protected := router.PathPrefix("/").Subrouter()
	protected.Use(func(next http.Handler) http.Handler {
		return authMware.Authenticate(next)
	})
	//protected.HandleFunc("/success", portalHandler.SuccessPage).Methods("GET") // Moved to public for testing

	// API routes
	api := router.PathPrefix("/api").Subrouter()
	api.Use(func(next http.Handler) http.Handler {
		return authMware.Authenticate(next)
	})
	api.HandleFunc("/authorize", apiHandler.HandleAuthorize).Methods("POST")
	api.HandleFunc("/accounting", apiHandler.HandleAccounting).Methods("POST")
	api.HandleFunc("/session", apiHandler.GetSession).Methods("GET")

	// Apply middleware - combine routers
	// Use a custom handler that routes logout separately (no CSRF)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle logout without CSRF
		if r.URL.Path == "/logout" {
			logoutRouter.ServeHTTP(w, r)
			return
		}
		// All other routes go through CSRF protection
		csrfMiddleware(router).ServeHTTP(w, r)
	})

	// Setup HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.App.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		appLogger.Info("Server starting", "port", cfg.App.Port, "env", cfg.App.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Server failed to start", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Server shutting down...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Fatal("Server forced to shutdown", "error", err)
	}

	appLogger.Info("Server stopped gracefully")
}
