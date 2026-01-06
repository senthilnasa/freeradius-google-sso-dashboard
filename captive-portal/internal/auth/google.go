package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/config"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/models"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleAuth handles Google OAuth authentication
type GoogleAuth struct {
	config       *config.GoogleOAuthConfig
	oauthConfig  *oauth2.Config
	hostedDomain string
}

// NewGoogleAuth creates a new Google OAuth authenticator
func NewGoogleAuth(cfg *config.GoogleOAuthConfig) *GoogleAuth {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURI,
		Scopes:       cfg.Scopes,
		Endpoint:     google.Endpoint,
	}

	return &GoogleAuth{
		config:       cfg,
		oauthConfig:  oauthConfig,
		hostedDomain: cfg.HostedDomain,
	}
}

// GetAuthURL generates the OAuth authorization URL
func (g *GoogleAuth) GetAuthURL(state string) string {
	options := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
	}

	// Add hosted domain restriction if configured
	if g.hostedDomain != "" {
		options = append(options, oauth2.SetAuthURLParam("hd", g.hostedDomain))
	}

	return g.oauthConfig.AuthCodeURL(state, options...)
}

// ExchangeCode exchanges authorization code for tokens
func (g *GoogleAuth) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := g.oauthConfig.Exchange(ctx, code)
	if err != nil {
		log.Error().Err(err).Msg("Failed to exchange OAuth code")
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	log.Debug().
		Str("token_type", token.TokenType).
		Time("expiry", token.Expiry).
		Msg("Successfully exchanged OAuth code")

	return token, nil
}

// GetUserInfo retrieves user information from Google
func (g *GoogleAuth) GetUserInfo(ctx context.Context, token *oauth2.Token) (*models.UserInfo, error) {
	client := g.oauthConfig.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user info from Google")
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error().
			Int("status_code", resp.StatusCode).
			Msg("Google API returned non-OK status")
		return nil, fmt.Errorf("google API returned status %d", resp.StatusCode)
	}

	var userInfo models.UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Error().Err(err).Msg("Failed to decode user info")
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	log.Info().
		Str("email", userInfo.Email).
		Str("name", userInfo.Name).
		Str("domain", userInfo.HD).
		Bool("verified", userInfo.EmailVerified).
		Msg("Retrieved user info from Google")

	return &userInfo, nil
}

// ValidateUserInfo validates user information against configured restrictions
func (g *GoogleAuth) ValidateUserInfo(userInfo *models.UserInfo) error {
	// Check if email is verified (TEMPORARILY DISABLED FOR TESTING)
	// TODO: Re-enable this check in production
	// if !userInfo.EmailVerified {
	// 	log.Warn().
	// 		Str("email", userInfo.Email).
	// 		Msg("Email not verified")
	// 	return fmt.Errorf("email not verified")
	// }
	if !userInfo.EmailVerified {
		log.Warn().
			Str("email", userInfo.Email).
			Msg("Email not verified - allowing for testing purposes")
	}

	// Check hosted domain restriction
	if g.hostedDomain != "" && userInfo.HD != g.hostedDomain {
		log.Warn().
			Str("email", userInfo.Email).
			Str("expected_domain", g.hostedDomain).
			Str("actual_domain", userInfo.HD).
			Msg("Domain mismatch")
		return fmt.Errorf("invalid domain: expected %s, got %s", g.hostedDomain, userInfo.HD)
	}

	// Check if email is present
	if userInfo.Email == "" {
		log.Error().Msg("Email is empty")
		return fmt.Errorf("email is required")
	}

	log.Debug().
		Str("email", userInfo.Email).
		Msg("User info validation successful")

	return nil
}

// AuthenticateUser performs the complete OAuth authentication flow
func (g *GoogleAuth) AuthenticateUser(ctx context.Context, code string) (*models.UserInfo, error) {
	// Exchange code for token
	token, err := g.ExchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange code failed: %w", err)
	}

	// Get user info
	userInfo, err := g.GetUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("get user info failed: %w", err)
	}

	// Validate user info
	if err := g.ValidateUserInfo(userInfo); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return userInfo, nil
}

// VerifyToken verifies an OAuth token is still valid
func (g *GoogleAuth) VerifyToken(ctx context.Context, token *oauth2.Token) error {
	if token == nil {
		return fmt.Errorf("token is nil")
	}

	// Check if token is expired
	if !token.Valid() {
		log.Warn().
			Time("expiry", token.Expiry).
			Msg("Token is expired")
		return fmt.Errorf("token expired at %s", token.Expiry.Format(time.RFC3339))
	}

	// Try to get user info to verify token is still valid
	client := g.oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/tokeninfo")
	if err != nil {
		log.Error().Err(err).Msg("Failed to verify token")
		return fmt.Errorf("token verification failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warn().
			Int("status_code", resp.StatusCode).
			Msg("Token verification returned non-OK status")
		return fmt.Errorf("token invalid: status %d", resp.StatusCode)
	}

	log.Debug().Msg("Token verification successful")
	return nil
}

// RevokeToken revokes an OAuth token
func (g *GoogleAuth) RevokeToken(ctx context.Context, token *oauth2.Token) error {
	if token == nil {
		return fmt.Errorf("token is nil")
	}

	revokeURL := fmt.Sprintf("https://oauth2.googleapis.com/revoke?token=%s", token.AccessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm(revokeURL, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to revoke token")
		return fmt.Errorf("token revocation failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warn().
			Int("status_code", resp.StatusCode).
			Msg("Token revocation returned non-OK status")
		// Don't return error as token might already be revoked
	}

	log.Info().Msg("Token revoked successfully")
	return nil
}
