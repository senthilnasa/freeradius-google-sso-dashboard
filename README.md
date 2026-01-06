# FreeRADIUS Google SSO Captive Portal with Dynamic VLAN Assignment

A production-ready, containerized captive portal solution that integrates **Google Workspace SSO** with **FreeRADIUS 3.x** for secure network authentication and **dynamic VLAN assignment** based on user email domain and attributes.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Docker](https://img.shields.io/badge/docker-ready-brightgreen.svg)
![FreeRADIUS](https://img.shields.io/badge/FreeRADIUS-3.x-orange.svg)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)
![PHP](https://img.shields.io/badge/PHP-8.x-777BB4.svg)

## 📋 Table of Contents

- [Features](#-features)
- [Architecture](#-architecture)
- [Prerequisites](#-prerequisites)
- [Quick Start](#-quick-start)
- [Configuration](#-configuration)
- [Documentation](#-documentation)
- [VLAN Assignment Logic](#-vlan-assignment-logic)
- [Admin Dashboard](#-admin-dashboard)
- [Security](#-security)
- [Troubleshooting](#-troubleshooting)
- [Contributing](#-contributing)
- [License](#-license)

## ✨ Features

### Authentication & Authorization
- 🔐 **Google Workspace SSO** integration via OAuth 2.0
- 📧 **Email-based VLAN assignment** with domain and suffix matching
- 👥 **Multi-domain support** with configurable rules
- 🔄 **Dynamic VLAN allocation** via RADIUS attributes
- ⚡ **Session management** with configurable timeout
- 🛡️ **Security hardening** with CSRF protection, XSS prevention

### RADIUS Integration
- 📡 **FreeRADIUS 3.x** with MySQL backend
- 🌐 **802.1X support** for wired and wireless networks
- 📊 **RADIUS accounting** for session tracking
- 🏷️ **Dynamic VLAN assignment** via Tunnel attributes
- 🔁 **Real-time accounting** updates

### Admin Dashboard
- 📈 **Comprehensive reporting** with multiple views
- 👤 **Role-based access control** (Superadmin, Netadmin, Helpdesk)
- 📊 **Authentication analytics** and trends
- 🔍 **Failed login monitoring** and alerts
- 📤 **Export functionality** (CSV/Excel)
- 🎯 **VLAN usage statistics**
- 🔐 **Audit trail** for all admin actions
- 👥 **User management** with password policies

### Infrastructure
- 🐳 **Fully containerized** with Docker Compose
- 🔒 **Automatic SSL** with Let's Encrypt (Certbot)
- 🚀 **Nginx reverse proxy** with rate limiting
- 💾 **MySQL 8.x** with optimized indexes
- 📝 **Structured logging** and monitoring
- 🔄 **Auto-renewal** for SSL certificates
- 📊 **Health checks** for all services
- 🎯 **Scalable architecture** (10,000+ concurrent users)

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                       User Device                           │
│                  (Wired/Wireless Client)                    │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ├─ Network Access Request
                        │
┌───────────────────────▼─────────────────────────────────────┐
│                  Network Switch/AP                          │
│               (802.1X / Captive Portal)                     │
└─────────┬────────────────────────────────────────┬──────────┘
          │                                        │
          │ RADIUS Auth/Acct                      │ HTTPS
          │                                        │
┌─────────▼───────────┐              ┌────────────▼──────────┐
│  FreeRADIUS 3.x     │◄─────────────│  Nginx + Certbot      │
│  - Dynamic VLAN     │   REST API   │  - Let's Encrypt SSL  │
│  - MySQL Backend    │              │  - Rate Limiting      │
└─────────┬───────────┘              └────────────┬──────────┘
          │                                       │
          │                       ┌───────────────┴──────────┐
          │                       │                          │
          │              ┌────────▼────────┐      ┌──────────▼────────┐
          │              │ Captive Portal  │      │  Admin Dashboard  │
          │              │   (Golang)      │      │      (PHP)        │
          │              │ - Google OAuth  │      │ - Reports         │
          │              │ - VLAN Logic    │      │ - User Mgmt       │
          │              └────────┬────────┘      └──────────┬────────┘
          │                       │                          │
          └───────────────────────┴──────────────────────────┘
                                  │
                       ┌──────────▼──────────┐
                       │    MySQL 8.x        │
                       │  - Sessions         │
                       │  - Accounting       │
                       │  - VLAN Rules       │
                       │  - Audit Logs       │
                       └─────────────────────┘
```

For detailed architecture, see [ARCHITECTURE.md](docs/ARCHITECTURE.md)

## 📦 Prerequisites

- **Docker Engine** 20.10+ and **Docker Compose** 2.0+
- **Domain names** with DNS configured (for SSL certificates)
- **Google Workspace** account with OAuth credentials
- **Linux server** with 4+ CPU cores and 8GB+ RAM
- **Network access** to RADIUS clients (switches/APs)

## 🚀 Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/senthilnasa/freeradius-google-sso-dashboard.git
cd freeradius-google-sso-dashboard
```

### 2. Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit configuration (see Configuration section)
nano .env
```

### 3. Setup Google OAuth

1. Create OAuth credentials in [Google Cloud Console](https://console.cloud.google.com/)
2. Configure authorized redirect URI: `https://portal.yourdomain.com/callback`
3. Add Client ID and Secret to `.env`

### 4. Initialize SSL Certificates

```bash
# Make script executable
chmod +x scripts/init-letsencrypt.sh

# Run SSL setup (requires DNS configured)
./scripts/init-letsencrypt.sh
```

### 5. Deploy Services

```bash
# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f
```

### 6. Access Dashboards

- **Captive Portal**: `https://portal.yourdomain.com`
- **Admin Dashboard**: `https://admin.yourdomain.com`
  - Default admin credentials (change immediately!):
    - Username: `admin`
    - Password: `admin123`

For detailed deployment instructions, see [DEPLOYMENT.md](docs/DEPLOYMENT.md)

## ⚙️ Configuration

### Required Environment Variables

```bash
# Google OAuth
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URI=https://portal.yourdomain.com/callback
GOOGLE_HOSTED_DOMAIN=yourdomain.com

# Domains
PORTAL_DOMAIN=portal.yourdomain.com
ADMIN_DOMAIN=admin.yourdomain.com

# Security Secrets (generate with: openssl rand -hex 32)
JWT_SECRET=your-32-char-minimum-secret
CSRF_SECRET=your-32-char-minimum-secret
SESSION_SECRET=your-32-char-minimum-secret
RADIUS_SECRET=your-radius-shared-secret

# Database
MYSQL_ROOT_PASSWORD=your-secure-password
MYSQL_PASSWORD=your-secure-password

# SSL
CERTBOT_EMAIL=admin@yourdomain.com
```

See [.env.example](.env.example) for complete configuration options.

## 📚 Documentation

- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)** - System architecture and design
- **[DEPLOYMENT.md](docs/DEPLOYMENT.md)** - Complete deployment guide
- **[ACCESS-POINT-CONFIGURATION.md](docs/ACCESS-POINT-CONFIGURATION.md)** - Configure Ruckus, Aruba, and UniFi APs
- **[IMPLEMENTATION-STATUS.md](docs/IMPLEMENTATION-STATUS.md)** - Project progress
- **API Documentation** - Coming soon
- **Operations Manual** - Coming soon

### Access Point Configuration

The system supports major enterprise AP vendors. See [ACCESS-POINT-CONFIGURATION.md](docs/ACCESS-POINT-CONFIGURATION.md) for detailed setup guides:

- **Ruckus** - SmartZone / ZoneDirector / Unleashed
- **Aruba** - Mobility Controller / Instant
- **UniFi** - Network Controller

Each guide includes:
- Step-by-step RADIUS configuration
- SSID/WLAN creation with WPA2/WPA3 Enterprise
- Dynamic VLAN assignment setup
- Testing and troubleshooting procedures

## 🎯 VLAN Assignment Logic

The system assigns VLANs based on email domain and username patterns:

### Configuration Example

```json
[
  {"domain":"krea.ac.in", "key":".mba", "Type":"Student-MBA", "VLAN":"216"},
  {"domain":"krea.ac.in", "key":".phd", "Type":"Student-PhD", "VLAN":"232"},
  {"domain":"krea.ac.in", "key":"", "Type":"Student-Others", "VLAN":"144"},
  {"domain":"krea.edu.in", "key":"", "Type":"Staff-Krea", "VLAN":"248"}
]
```

### Assignment Rules

1. **Email parsing**: Extract domain and username from email
2. **Domain matching**: Find rules for the user's domain
3. **Suffix matching**: Check if username contains key pattern (e.g., `.mba`)
4. **Priority-based**: Evaluate rules by priority (lower = higher priority)
5. **Default fallback**: Use empty key `""` as default for domain

### Example

- `john.mba@krea.ac.in` → VLAN 216 (Student-MBA)
- `jane.phd@krea.ac.in` → VLAN 232 (Student-PhD)
- `alice@krea.ac.in` → VLAN 144 (Student-Others - default)
- `bob@krea.edu.in` → VLAN 248 (Staff-Krea)

## 📊 Admin Dashboard

### User Roles

| Role | Permissions |
|------|-------------|
| **Superadmin** | Full system access, user management, all reports |
| **Netadmin** | User management, all reports, VLAN configuration |
| **Helpdesk** | Read-only access to reports and logs |

### Reports

1. **Daily Authentication Summary**
   - Total attempts, success/failure counts
   - Hourly breakdown
   - VLAN distribution
   - Error type analysis

2. **Monthly Usage Report**
   - Daily session statistics
   - Data usage tracking
   - Unique user counts

3. **Failed Login Report**
   - Failed authentication patterns
   - Threshold-based filtering
   - Error type grouping

4. **User Type Distribution**
   - Authentication patterns by user type
   - VLAN correlation analysis

5. **System Health**
   - Database statistics
   - Active sessions
   - NAS device activity

## 🔒 Security

### Authentication Security
- ✅ OAuth 2.0 with Google Workspace
- ✅ Domain restriction enforcement
- ✅ Session timeout and idle timeout
- ✅ CSRF token validation
- ✅ JWT-based session tokens

### Network Security
- ✅ TLS 1.2+ enforcement
- ✅ HTTPS-only (HSTS enabled)
- ✅ Rate limiting on all endpoints
- ✅ Secure cookie flags
- ✅ Security headers (CSP, X-Frame-Options, etc.)

### Data Security
- ✅ SQL injection prevention (prepared statements)
- ✅ XSS protection
- ✅ Password hashing (bcrypt)
- ✅ Secrets management via environment variables
- ✅ Audit logging

### Infrastructure Security
- ✅ Container isolation
- ✅ Non-root containers
- ✅ Minimal base images
- ✅ Regular security updates
- ✅ Health monitoring

## 🐛 Troubleshooting

### Common Issues

**Services won't start:**
```bash
# Check logs
docker-compose logs

# Restart services
docker-compose restart
```

**SSL certificate issues:**
```bash
# Check certificate status
docker-compose exec certbot certificates

# Test renewal
docker-compose run --rm certbot renew --dry-run
```

**Authentication failures:**
```bash
# Check FreeRADIUS logs
docker-compose logs freeradius

# Check database
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} radius
```

**VLAN not assigned:**
```bash
# Verify VLAN rules
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} radius \
  -e "SELECT * FROM vlan_rules WHERE enabled=1 ORDER BY priority;"

# Check session logs
docker-compose logs captive-portal | grep VLAN
```

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [FreeRADIUS](https://freeradius.org/) - The open source RADIUS server
- [Let's Encrypt](https://letsencrypt.org/) - Free SSL/TLS certificates
- [Google OAuth](https://developers.google.com/identity/protocols/oauth2) - Authentication provider

## 📞 Support

- 🐛 Issues: [GitHub Issues](https://github.com/senthilnasa/freeradius-google-sso-dashboard/issues)
- 📖 Documentation: [docs/](docs/)
- 🔗 Repository: [GitHub](https://github.com/senthilnasa/freeradius-google-sso-dashboard)

---

**Made with ❤️ for secure network authentication**
