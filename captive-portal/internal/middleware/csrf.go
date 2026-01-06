package middleware

import (
	"net/http"

	"github.com/gorilla/csrf"
)

// SetupCSRF creates CSRF protection middleware
func SetupCSRF(secret string, secure bool) func(http.Handler) http.Handler {
	return csrf.Protect(
		[]byte(secret),
		csrf.Secure(secure),
		csrf.Path("/"),
		csrf.HttpOnly(true),
		csrf.SameSite(csrf.SameSiteStrictMode),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "CSRF token validation failed", http.StatusForbidden)
		})),
	)
}
