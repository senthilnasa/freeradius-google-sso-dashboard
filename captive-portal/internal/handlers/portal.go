package handlers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/models"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/internal/session"
	"github.com/senthilnasa/freeradius-google-sso-dashboard/pkg/logger"
)

type PortalHandler struct {
	sessionService *session.Service
	logger         *logger.Logger
	logoURL        string
	institutionName string
	primaryColor   string
}

func NewPortalHandler(sessionService *session.Service, logger *logger.Logger, logoURL, institutionName, primaryColor string) *PortalHandler {
	return &PortalHandler{
		sessionService: sessionService,
		logger:         logger,
		logoURL:        logoURL,
		institutionName: institutionName,
		primaryColor:   primaryColor,
	}
}

func (h *PortalHandler) LandingPage(w http.ResponseWriter, r *http.Request) {
	// Check if user already has a valid session
	cookie, err := r.Cookie("auth_token")
	if err == nil && cookie.Value != "" {
		// Try to validate the token
		claims, err := h.sessionService.ValidateToken(cookie.Value)
		if err == nil {
			// Token is valid, check if session is active in database
			session, err := h.sessionService.Get(claims.SessionID)
			if err == nil && session != nil && session.IsActive {
				h.logger.Debug("User has active session, redirecting to success", "session_id", claims.SessionID)
				// User is already authenticated, redirect to success page
				http.Redirect(w, r, "/success", http.StatusSeeOther)
				return
			} else {
				// Session not active or not found - clear the cookie
				h.logger.Debug("Session inactive or not found, clearing cookie", "session_id", claims.SessionID)
				http.SetCookie(w, &http.Cookie{
					Name:     "auth_token",
					Value:    "",
					Path:     "/",
					HttpOnly: true,
					Secure:   false,
					MaxAge:   -1,
					SameSite: http.SameSiteLaxMode,
				})
			}
		} else {
			// Token validation failed - clear the cookie
			h.logger.Debug("Token validation failed, clearing cookie")
			http.SetCookie(w, &http.Cookie{
				Name:     "auth_token",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				Secure:   false,
				MaxAge:   -1,
				SameSite: http.SameSiteLaxMode,
			})
		}
	}

	// No valid session, show login page
	tmpl := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Network Authentication - %s</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 2rem 1rem;
            background: linear-gradient(135deg, %s 0%%, #764ba2 100%%);
        }
        .container {
            text-align: center;
            background: white;
            padding: 2.5rem;
            border-radius: 12px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            width: 100%%;
            max-width: 450px;
        }
        .logo {
            height: 60px;
            max-width: 100%%;
            object-fit: contain;
            display: block;
            margin: 0 auto 1.5rem;
        }
        h1 {
            color: #333;
            margin-bottom: 1rem;
            font-size: 1.75rem;
            font-weight: 600;
        }
        p {
            color: #666;
            margin-bottom: 2rem;
            font-size: 1rem;
            line-height: 1.5;
        }
        .btn {
            display: inline-block;
            padding: 14px 32px;
            background: #4285f4;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            font-weight: 600;
            font-size: 1rem;
            transition: all 0.3s ease;
            box-shadow: 0 4px 6px rgba(66, 133, 244, 0.3);
        }
        .btn:hover {
            background: #357ae8;
            box-shadow: 0 6px 12px rgba(66, 133, 244, 0.4);
            transform: translateY(-2px);
        }
        .btn:active {
            transform: translateY(0);
        }
        @media (max-width: 768px) {
            .logo {
                height: 50px;
            }
            .container {
                padding: 2rem 1.5rem;
            }
            h1 {
                font-size: 1.5rem;
            }
            p {
                font-size: 0.95rem;
            }
        }
        @media (max-width: 480px) {
            .logo {
                height: 45px;
            }
            .container {
                padding: 1.75rem 1.25rem;
            }
            h1 {
                font-size: 1.35rem;
            }
            .btn {
                padding: 12px 28px;
                font-size: 0.95rem;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <img src="%s" alt="%s" class="logo">
        <h1>Welcome to %s</h1>
        <p>Please sign in with your %s account to access the wifi network</p>
        <a href="/login" class="btn">Sign in with Google</a>
    </div>
</body>
</html>
`, h.institutionName, h.primaryColor, h.logoURL, h.institutionName, h.institutionName, h.institutionName)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(tmpl))
}

func (h *PortalHandler) SuccessPage(w http.ResponseWriter, r *http.Request) {
	// Try to get session from context (if on protected route)
	sess := r.Context().Value("session")
	var sessionData *models.Session

	if sess != nil {
		// Session available from middleware
		var ok bool
		sessionData, ok = sess.(*models.Session)
		if !ok {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	} else {
		// No session in context - try to get from cookie/token
		// Extract token from cookie
		cookie, err := r.Cookie("auth_token")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Validate token and get session
		claims, err := h.sessionService.ValidateToken(cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Get session from database
		sessionData, err = h.sessionService.Get(claims.SessionID)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	// Create template data with session and branding
	type TemplateData struct {
		Session         *models.Session
		LogoURL         string
		InstitutionName string
		PrimaryColor    string
	}

	data := TemplateData{
		Session:         sessionData,
		LogoURL:         h.logoURL,
		InstitutionName: h.institutionName,
		PrimaryColor:    h.primaryColor,
	}

	tmpl := template.Must(template.New("success").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Authentication Successful - {{.InstitutionName}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 2rem 1rem;
            background: linear-gradient(135deg, {{.PrimaryColor}} 0%, #764ba2 100%);
        }
        .container {
            background: white;
            padding: 2.5rem;
            border-radius: 12px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            width: 100%;
            max-width: 550px;
        }
        .logo {
            height: 60px;
            max-width: 100%;
            object-fit: contain;
            display: block;
            margin: 0 auto 1.5rem;
        }
        .success-icon {
            width: 64px;
            height: 64px;
            background: #28a745;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 1.5rem;
            color: white;
            font-size: 2rem;
        }
        h1 {
            color: #28a745;
            margin-bottom: 0.5rem;
            font-size: 1.75rem;
            font-weight: 600;
            text-align: center;
        }
        .subtitle {
            color: #666;
            margin-bottom: 2rem;
            font-size: 1rem;
            text-align: center;
        }
        .info {
            background: #f8f9fa;
            padding: 1.5rem;
            border-radius: 8px;
            margin-bottom: 1.5rem;
        }
        .info-item {
            margin: 0.75rem 0;
            display: flex;
            justify-content: space-between;
            align-items: center;
            flex-wrap: wrap;
            gap: 0.5rem;
        }
        .label {
            font-weight: 600;
            color: #495057;
            font-size: 0.95rem;
        }
        .value {
            color: #212529;
            font-size: 0.95rem;
            word-break: break-all;
        }
        .btn {
            display: block;
            width: 100%;
            padding: 12px;
            background: #dc3545;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            border: none;
            cursor: pointer;
            font-size: 1rem;
            font-weight: 600;
            transition: all 0.3s ease;
            text-align: center;
        }
        .btn:hover {
            background: #c82333;
            transform: translateY(-2px);
            box-shadow: 0 4px 8px rgba(220, 53, 69, 0.3);
        }
        .btn:active {
            transform: translateY(0);
        }
        @media (max-width: 768px) {
            .logo {
                height: 50px;
            }
            .container {
                padding: 2rem 1.5rem;
            }
            h1 {
                font-size: 1.5rem;
            }
            .success-icon {
                width: 56px;
                height: 56px;
                font-size: 1.75rem;
            }
        }
        @media (max-width: 480px) {
            .logo {
                height: 45px;
            }
            .container {
                padding: 1.75rem 1.25rem;
            }
            h1 {
                font-size: 1.35rem;
            }
            .info {
                padding: 1.25rem;
            }
            .info-item {
                flex-direction: column;
                align-items: flex-start;
            }
            .btn {
                padding: 11px;
                font-size: 0.95rem;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <img src="{{.LogoURL}}" alt="{{.InstitutionName}}" class="logo">
        <div class="success-icon">✓</div>
        <h1>Authentication Successful</h1>
        <p class="subtitle">You are now connected to the network</p>
        <div class="info">
            <div class="info-item">
                <span class="label">Email:</span>
                <span class="value">{{.Session.Email}}</span>
            </div>
            <div class="info-item">
                <span class="label">User Type:</span>
                <span class="value">{{.Session.UserType}}</span>
            </div>
            <div class="info-item">
                <span class="label">IP Address:</span>
                <span class="value">{{.Session.IPAddress}}</span>
            </div>
        </div>
        <a href="/logout" class="btn" style="text-decoration: none; display: block;">Logout</a>
    </div>
</body>
</html>
`))

	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, data)
}

func (h *PortalHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}
