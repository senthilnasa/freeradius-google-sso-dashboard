# System Architecture

**Project Repository:** https://github.com/senthilnasa/freeradius-google-sso-dashboard

## Overview

This system provides a secure captive portal solution using Google Workspace SSO integrated with FreeRADIUS for dynamic VLAN assignment based on email domain and suffix rules.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         User Device                             │
│                    (Wired/Wireless Client)                      │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ├─ Network Access Request
                            │
┌───────────────────────────▼─────────────────────────────────────┐
│                      Network Switch/AP                          │
│                   (802.1X / Captive Portal)                     │
└─────────┬─────────────────────────────────────────┬─────────────┘
          │                                         │
          │ RADIUS Auth/Acct                       │ HTTP Redirect
          │                                         │
┌─────────▼───────────┐                   ┌─────────▼─────────────┐
│   FreeRADIUS 3.x    │◄──────────────────│   Nginx (HTTPS)       │
│   - AAA Server      │   REST API        │   - Reverse Proxy     │
│   - MySQL Backend   │                   │   - TLS Termination   │
│   - Dynamic VLAN    │                   └─────────┬─────────────┘
│   - Accounting      │                             │
└─────────┬───────────┘                             │
          │                          ┌──────────────┴──────────────┐
          │                          │                             │
          │                  ┌───────▼────────┐          ┌─────────▼────────┐
          │                  │ Captive Portal │          │  Admin Dashboard │
          │                  │   (Golang)     │          │      (PHP)       │
          │                  │ - Google OAuth │          │ - Reporting      │
          │                  │ - VLAN Logic   │          │ - Monitoring     │
          │                  │ - Session Mgmt │          │ - Analytics      │
          │                  └───────┬────────┘          └─────────┬────────┘
          │                          │                             │
          └──────────────────────────┴─────────────────────────────┘
                                     │
                          ┌──────────▼──────────┐
                          │    MySQL 8.x        │
                          │  - User Data        │
                          │  - Sessions         │
                          │  - Accounting       │
                          │  - VLAN Rules       │
                          │  - Audit Logs       │
                          └─────────────────────┘
```

## Component Details

### 1. FreeRADIUS Container
- **Purpose**: AAA (Authentication, Authorization, Accounting)
- **Technology**: FreeRADIUS 3.x
- **Key Features**:
  - MySQL backend for user data
  - Dynamic VLAN assignment via RADIUS attributes
  - Session accounting (radacct table)
  - Authentication logging (radpostauth table)
  - REST API integration with captive portal
- **RADIUS Attributes**:
  - `Tunnel-Type = VLAN (13)`
  - `Tunnel-Medium-Type = IEEE-802 (6)`
  - `Tunnel-Private-Group-Id = <VLAN_ID>`

### 2. Golang Captive Portal
- **Purpose**: User authentication and VLAN assignment logic
- **Technology**: Go 1.21+
- **Key Features**:
  - Google OAuth 2.0 integration
  - Email domain and suffix validation
  - VLAN decision engine
  - Session management (JWT-based)
  - RADIUS authorization trigger
  - REST API for FreeRADIUS callbacks
  - Security: CSRF, XSS protection, secure cookies
- **Endpoints**:
  - `GET /` - Landing page
  - `GET /login` - Initiate Google OAuth
  - `GET /callback` - OAuth callback handler
  - `POST /authorize` - RADIUS authorization endpoint
  - `POST /accounting` - RADIUS accounting endpoint
  - `GET /logout` - Session termination

### 3. PHP Admin Dashboard
- **Purpose**: Reporting, monitoring, and analytics
- **Technology**: PHP 8.x
- **Key Features**:
  - Admin authentication
  - User session tracking
  - VLAN usage reports
  - Failed login analytics
  - Export functionality (CSV/Excel)
  - Date/domain/VLAN filtering
  - Real-time session monitoring
- **Pages**:
  - `/admin/login` - Admin login
  - `/admin/dashboard` - Overview
  - `/admin/sessions` - Active sessions
  - `/admin/reports` - Historical reports
  - `/admin/vlan-config` - VLAN rule management
  - `/admin/audit` - Audit logs

### 4. MySQL Database
- **Purpose**: Central data store
- **Technology**: MySQL 8.x
- **Schema**:
  - `radcheck` - User credentials
  - `radreply` - User-specific RADIUS replies
  - `radacct` - Accounting/session data
  - `radpostauth` - Authentication logs
  - `vlan_rules` - Domain/suffix VLAN mappings
  - `sessions` - Active user sessions
  - `admin_users` - Dashboard admin accounts
  - `audit_logs` - System audit trail

### 5. Nginx Reverse Proxy
- **Purpose**: TLS termination and routing
- **Technology**: Nginx
- **Configuration**:
  - HTTPS enforcement
  - Reverse proxy to captive portal (port 8080)
  - Reverse proxy to admin dashboard (port 9000)
  - Security headers
  - Rate limiting
  - Static file serving

## Data Flow

### Authentication Flow

```
1. User connects to network
   ↓
2. Network device initiates captive portal redirect
   ↓
3. User browser loads captive portal landing page
   ↓
4. User clicks "Login with Google"
   ↓
5. Captive portal redirects to Google OAuth
   ↓
6. User authenticates with Google
   ↓
7. Google redirects back with authorization code
   ↓
8. Captive portal exchanges code for access token
   ↓
9. Captive portal retrieves user email from Google
   ↓
10. System validates email domain and suffix
    ↓
11. VLAN assignment logic evaluates rules
    ↓
12. Captive portal triggers RADIUS authorization
    ↓
13. FreeRADIUS stores user session and VLAN
    ↓
14. Network device receives Access-Accept with VLAN attributes
    ↓
15. User is assigned to appropriate VLAN
    ↓
16. User gains network access
```

### VLAN Assignment Logic

```go
// Pseudocode
func assignVLAN(email string) (vlanID string, userType string) {
    domain := extractDomain(email)      // e.g., "krea.ac.in"
    username := extractUsername(email)  // e.g., "john.mba"

    rules := loadVLANRules() // From database

    for _, rule := range rules {
        if rule.Domain == domain {
            if rule.Key == "" {
                // Default rule for domain
                return rule.VLAN, rule.Type
            }
            if strings.Contains(username, rule.Key) {
                // Suffix match (e.g., ".mba")
                return rule.VLAN, rule.Type
            }
        }
    }

    // No match found - deny access or use global default
    return "", ""
}
```

## Security Considerations

### 1. Authentication & Authorization
- Google OAuth 2.0 with Workspace domain restriction
- JWT-based session tokens with expiry
- RADIUS shared secret protection (environment variables)
- Admin dashboard password hashing (bcrypt)

### 2. Transport Security
- TLS 1.2+ enforcement
- HTTPS for all web traffic
- Secure cookie flags (HttpOnly, Secure, SameSite)

### 3. Input Validation
- SQL injection prevention (prepared statements)
- XSS protection (input sanitization, CSP headers)
- CSRF token validation

### 4. Network Security
- Rate limiting on authentication endpoints
- Session timeout (30 minutes idle, 8 hours absolute)
- Failed login attempt tracking

### 5. Audit & Logging
- All authentication attempts logged
- VLAN assignment decisions logged
- Admin actions audited
- Centralized logging to MySQL

## Scalability

### Performance Targets
- Support 10,000+ concurrent users
- Sub-second authentication response time
- High availability with container orchestration ready

### Database Optimization
- Indexed columns: email, session_id, timestamp, vlan_id
- Partitioning for radacct table (by date)
- Connection pooling in Golang and PHP

### Caching Strategy
- VLAN rules cached in memory with TTL
- Session data cached (Redis optional for future)

## Deployment

### Docker Compose Services
1. `mysql` - Database
2. `freeradius` - RADIUS server
3. `captive-portal` - Golang app
4. `admin-dashboard` - PHP app
5. `nginx` - Reverse proxy

### Environment Variables
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`
- `GOOGLE_REDIRECT_URI`
- `MYSQL_ROOT_PASSWORD`
- `MYSQL_DATABASE`
- `MYSQL_USER`
- `MYSQL_PASSWORD`
- `RADIUS_SECRET`
- `JWT_SECRET`
- `SESSION_TIMEOUT`

### Volumes
- `mysql-data` - Database persistence
- `freeradius-config` - RADIUS configuration
- `ssl-certs` - TLS certificates
- `logs` - Application logs

## Monitoring & Maintenance

### Health Checks
- FreeRADIUS: `radtest` command
- Captive Portal: `/health` endpoint
- Admin Dashboard: `/health` endpoint
- MySQL: Connection test

### Metrics to Monitor
- Active sessions count
- Authentication success/failure rate
- VLAN distribution
- Database connection pool usage
- API response times

### Log Rotation
- Application logs rotated daily
- Database logs managed by MySQL
- Retention: 90 days

## Future Enhancements
- Multi-factor authentication (MFA)
- Redis for session caching
- Kubernetes deployment manifests
- Prometheus metrics export
- Grafana dashboards
- Email notifications for admin events
- Self-service password reset
- Mobile app integration
