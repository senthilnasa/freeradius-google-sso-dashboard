# Deployment Guide

## FreeRADIUS Google SSO Captive Portal with Dynamic VLAN

This guide walks you through deploying the complete system in a production environment.

---

## Prerequisites

### System Requirements
- **Operating System**: Linux (Ubuntu 20.04+ or CentOS 8+ recommended)
- **CPU**: Minimum 4 cores (8+ recommended for production)
- **RAM**: Minimum 8GB (16GB+ recommended for 10,000+ users)
- **Storage**: Minimum 50GB SSD (for logs and database)
- **Network**: Static IP address with proper DNS configuration

### Software Requirements
- Docker Engine 20.10+
- Docker Compose 2.0+
- Git
- OpenSSL

### Network Requirements
- **Ports to open**:
  - 80 (HTTP - for Let's Encrypt validation)
  - 443 (HTTPS - for web access)
  - 1812/UDP (RADIUS Authentication)
  - 1813/UDP (RADIUS Accounting)
- **DNS**: Both portal and admin domains must resolve to server IP
- **Firewall**: Allow incoming connections on above ports

---

## Step 1: Clone Repository

```bash
# Clone the repository
git clone https://github.com/yourusername/freeradius-google-sso-dashboard.git
cd freeradius-google-sso-dashboard

# Make scripts executable
chmod +x scripts/*.sh
```

---

## Step 2: Configure DNS

Before proceeding, ensure your domains are properly configured:

```bash
# Example DNS configuration (in your domain registrar/DNS provider):
# A Record: portal.yourdomain.com -> YOUR_SERVER_IP
# A Record: admin.yourdomain.com  -> YOUR_SERVER_IP

# Verify DNS resolution:
nslookup portal.yourdomain.com
nslookup admin.yourdomain.com
```

**⚠️ Important**: DNS must be configured BEFORE requesting SSL certificates!

---

## Step 3: Google OAuth Setup

### 3.1 Create Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable "Google+ API" or "People API"

### 3.2 Configure OAuth Consent Screen

1. Navigate to "APIs & Services" → "OAuth consent screen"
2. Select "Internal" for Google Workspace users only
3. Fill in application name and support email
4. Add authorized domains (your domain)
5. Save

### 3.3 Create OAuth Credentials

1. Navigate to "APIs & Services" → "Credentials"
2. Click "Create Credentials" → "OAuth 2.0 Client ID"
3. Application type: "Web application"
4. Name: "RADIUS Captive Portal"
5. Authorized JavaScript origins:
   - `https://portal.yourdomain.com`
6. Authorized redirect URIs:
   - `https://portal.yourdomain.com/callback`
7. Click "Create"
8. **Save the Client ID and Client Secret** (you'll need these)

---

## Step 4: Environment Configuration

### 4.1 Create Environment File

```bash
# Copy example environment file
cp .env.example .env

# Edit with your settings
nano .env
```

### 4.2 Required Configuration

Edit `.env` file with your values:

```bash
# Database
MYSQL_ROOT_PASSWORD=your-super-secure-root-password
MYSQL_PASSWORD=your-secure-mysql-password

# Google OAuth (from Step 3)
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URI=https://portal.yourdomain.com/callback
GOOGLE_HOSTED_DOMAIN=yourdomain.com

# RADIUS
RADIUS_SECRET=your-secure-radius-secret

# Security Secrets (generate using: openssl rand -hex 32)
JWT_SECRET=your-jwt-secret-minimum-32-characters
CSRF_SECRET=your-csrf-secret-minimum-32-characters
SESSION_SECRET=your-session-secret-minimum-32-characters
REDIS_PASSWORD=your-redis-password

# Domains
PORTAL_DOMAIN=portal.yourdomain.com
ADMIN_DOMAIN=admin.yourdomain.com
APP_URL=https://portal.yourdomain.com
ADMIN_URL=https://admin.yourdomain.com

# SSL/TLS
CERTBOT_EMAIL=admin@yourdomain.com
CERTBOT_PRODUCTION=true

# Application
APP_ENV=production
LOG_LEVEL=info
```

### 4.3 Generate Secure Secrets

```bash
# Generate secure random strings for secrets
echo "JWT_SECRET=$(openssl rand -hex 32)"
echo "CSRF_SECRET=$(openssl rand -hex 32)"
echo "SESSION_SECRET=$(openssl rand -hex 32)"
echo "RADIUS_SECRET=$(openssl rand -hex 16)"
echo "MYSQL_ROOT_PASSWORD=$(openssl rand -hex 16)"
echo "MYSQL_PASSWORD=$(openssl rand -hex 16)"
```

---

## Step 5: SSL Certificate Setup

### 5.1 Initialize Let's Encrypt

We provide a script to automate SSL certificate setup:

```bash
# Run the Let's Encrypt initialization script
./scripts/init-letsencrypt.sh
```

The script will:
1. Validate your configuration
2. Check DNS resolution
3. Create temporary self-signed certificates
4. Start Nginx
5. Request Let's Encrypt certificates
6. Configure automatic renewal

### 5.2 Manual Certificate Setup (Alternative)

If you prefer manual setup:

```bash
# Start services without SSL first
docker-compose up -d mysql freeradius captive-portal admin-dashboard

# Start Nginx
docker-compose up -d nginx

# Request certificates for portal
docker-compose run --rm certbot certonly \
  --webroot \
  --webroot-path=/var/www/certbot \
  --email admin@yourdomain.com \
  --agree-tos \
  --no-eff-email \
  -d portal.yourdomain.com

# Request certificates for admin
docker-compose run --rm certbot certonly \
  --webroot \
  --webroot-path=/var/www/certbot \
  --email admin@yourdomain.com \
  --agree-tos \
  --no-eff-email \
  -d admin.yourdomain.com

# Reload Nginx
docker-compose exec nginx nginx -s reload

# Start renewal service
docker-compose up -d certbot-renew
```

### 5.3 Using Custom SSL Certificates

If you have your own SSL certificates:

```bash
# Copy your certificates to nginx/ssl/
cp your-portal.crt nginx/ssl/portal.yourdomain.com.crt
cp your-portal.key nginx/ssl/portal.yourdomain.com.key
cp your-admin.crt nginx/ssl/admin.yourdomain.com.crt
cp your-admin.key nginx/ssl/admin.yourdomain.com.key

# Update nginx configuration to use custom paths
# Edit nginx/conf.d/captive-portal.conf and admin-dashboard.conf
```

---

## Step 6: Deploy Services

### 6.1 Start All Services

```bash
# Start all services
docker-compose up -d

# Check service status
docker-compose ps

# View logs
docker-compose logs -f
```

### 6.2 Verify Services

```bash
# Check MySQL
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} -e "SHOW DATABASES;"

# Check FreeRADIUS
docker-compose exec freeradius radtest test test localhost 0 ${RADIUS_SECRET}

# Check Captive Portal
curl -k https://portal.yourdomain.com/health

# Check Admin Dashboard
curl -k https://admin.yourdomain.com/health.php

# Check Nginx
curl -I https://portal.yourdomain.com
```

---

## Step 7: Initial Configuration

### 7.1 Update Default Admin Passwords

```bash
# Connect to MySQL
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} radius

# Update admin user passwords (use your own secure passwords)
UPDATE admin_users SET password_hash = '$2y$10$NEW_BCRYPT_HASH_HERE' WHERE username = 'admin';
UPDATE admin_users SET password_hash = '$2y$10$NEW_BCRYPT_HASH_HERE' WHERE username = 'netadmin';
UPDATE admin_users SET password_hash = '$2y$10$NEW_BCRYPT_HASH_HERE' WHERE username = 'helpdesk';
```

To generate bcrypt hashes:
```bash
# Using PHP
php -r "echo password_hash('YourNewPassword', PASSWORD_BCRYPT);"

# Using Python
python3 -c "import bcrypt; print(bcrypt.hashpw(b'YourNewPassword', bcrypt.gensalt()).decode())"
```

### 7.2 Configure NAS Devices

Update RADIUS clients for your network devices:

```bash
# Edit NAS configuration
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} radius

# Add your network switches/access points
INSERT INTO nas (nasname, shortname, type, secret, description) VALUES
('192.168.1.10', 'core-switch-1', 'cisco', 'your-radius-secret', 'Core Switch 1'),
('192.168.1.11', 'core-switch-2', 'cisco', 'your-radius-secret', 'Core Switch 2'),
('192.168.1.20', 'wireless-controller', 'other', 'your-radius-secret', 'Wireless Controller');
```

### 7.3 Verify VLAN Rules

```bash
# Check VLAN rules
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} radius -e "SELECT * FROM vlan_rules ORDER BY priority;"

# Add custom rules if needed
INSERT INTO vlan_rules (domain, key, user_type, vlan_id, priority, enabled) VALUES
('yourdomain.com', '.student', 'Student', 100, 10, 1);
```

---

## Step 8: Testing

### 8.1 Test Authentication Flow

1. **Access Captive Portal**:
   ```
   https://portal.yourdomain.com
   ```

2. **Click "Login with Google"**

3. **Authenticate with Google Workspace account**

4. **Verify**:
   - Successful authentication
   - VLAN assignment
   - Session created
   - Logged in admin dashboard

### 8.2 Test RADIUS Authentication

```bash
# Test from FreeRADIUS container
docker-compose exec freeradius radtest testuser@yourdomain.com password localhost 0 ${RADIUS_SECRET}

# Expected: Access-Accept with Tunnel attributes
```

### 8.3 Test Admin Dashboard

1. Access: `https://admin.yourdomain.com`
2. Login with default credentials (change immediately!)
3. Verify all reports are accessible
4. Test export functionality

---

## Step 9: Network Device Configuration

### 9.1 Cisco Switch Example

```cisco
! Enable 802.1X
aaa new-model
aaa authentication login default local
aaa authentication dot1x default group radius
aaa authorization network default group radius

! Configure RADIUS server
radius server RADIUS-PRIMARY
 address ipv4 YOUR_SERVER_IP auth-port 1812 acct-port 1813
 key your-radius-secret
 timeout 5
 retransmit 3

! Enable dynamic VLAN assignment
aaa authorization network default group radius
vlan 100-250
 name DYNAMIC-VLANS

! Configure interface
interface GigabitEthernet1/0/1
 switchport mode access
 authentication port-control auto
 dot1x pae authenticator
 spanning-tree portfast
```

### 9.2 Wireless Controller Example

```
# Configure RADIUS authentication
config radius auth add 1 YOUR_SERVER_IP 1812 ascii your-radius-secret
config radius acct add 1 YOUR_SERVER_IP 1813 ascii your-radius-secret

# Enable dynamic VLAN
config wlan radius_server auth add WLAN_NAME 1
config wlan radius_server acct add WLAN_NAME 1
config wlan interface WLAN_NAME dynamic-interface enable
```

---

## Step 10: Monitoring & Maintenance

### 10.1 Log Locations

```bash
# Application logs
logs/captive-portal/
logs/admin-dashboard/
logs/freeradius/
logs/nginx/
certbot/logs/

# View logs
docker-compose logs -f captive-portal
docker-compose logs -f freeradius
docker-compose logs -f nginx
```

### 10.2 Database Backup

```bash
# Create backup script
cat > scripts/backup-database.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
docker-compose exec -T mysql mysqldump -u root -p${MYSQL_ROOT_PASSWORD} radius > ${BACKUP_DIR}/radius_${TIMESTAMP}.sql
gzip ${BACKUP_DIR}/radius_${TIMESTAMP}.sql
# Keep only last 30 days
find ${BACKUP_DIR} -name "radius_*.sql.gz" -mtime +30 -delete
EOF

chmod +x scripts/backup-database.sh

# Add to crontab
crontab -e
# Add: 0 2 * * * /path/to/scripts/backup-database.sh
```

### 10.3 Certificate Renewal

Certificates auto-renew via the `certbot-renew` container. Monitor:

```bash
# Check renewal logs
docker-compose logs certbot-renew

# Manual renewal test
docker-compose run --rm certbot renew --dry-run

# Force renewal
docker-compose run --rm certbot renew --force-renewal
```

### 10.4 Health Checks

```bash
# Create health check script
cat > scripts/health-check.sh << 'EOF'
#!/bin/bash
echo "=== Service Health Check ==="
echo "Portal: $(curl -s -o /dev/null -w '%{http_code}' https://portal.yourdomain.com/health)"
echo "Admin: $(curl -s -o /dev/null -w '%{http_code}' https://admin.yourdomain.com/health.php)"
echo "MySQL: $(docker-compose exec -T mysql mysqladmin ping -h localhost -u root -p${MYSQL_ROOT_PASSWORD} 2>&1 | grep -c 'alive')"
echo "FreeRADIUS: $(docker-compose exec -T freeradius radtest test test localhost 0 ${RADIUS_SECRET} 2>&1 | grep -c 'Access-')"
EOF

chmod +x scripts/health-check.sh
```

---

## Step 11: Security Hardening

### 11.1 Firewall Configuration

```bash
# Using UFW (Ubuntu)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 1812/udp
sudo ufw allow 1813/udp
sudo ufw enable

# Using firewalld (CentOS)
sudo firewall-cmd --permanent --add-port=80/tcp
sudo firewall-cmd --permanent --add-port=443/tcp
sudo firewall-cmd --permanent --add-port=1812/udp
sudo firewall-cmd --permanent --add-port=1813/udp
sudo firewall-cmd --reload
```

### 11.2 Restrict Database Access

```bash
# Update docker-compose.yml to remove MySQL port exposure
# Comment out:
# ports:
#   - "3306:3306"
```

### 11.3 Enable Audit Logging

Already enabled in the database schema. Monitor via admin dashboard.

---

## Troubleshooting

### Services Won't Start

```bash
# Check logs
docker-compose logs

# Check specific service
docker-compose logs freeradius

# Restart services
docker-compose restart
```

### SSL Certificate Issues

```bash
# Check certificate status
docker-compose exec certbot certificates

# Check Nginx configuration
docker-compose exec nginx nginx -t

# View Let's Encrypt logs
cat certbot/logs/letsencrypt.log
```

### Authentication Failures

```bash
# Check FreeRADIUS logs
docker-compose logs freeradius | grep "Auth: "

# Check VLAN assignment
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} radius -e "SELECT * FROM radpostauth ORDER BY authdate DESC LIMIT 10;"

# Test VLAN logic
docker-compose exec captive-portal /app/captive-portal test-vlan user@domain.com
```

### Database Connection Issues

```bash
# Check MySQL status
docker-compose ps mysql

# Test connection
docker-compose exec mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} -e "SELECT 1;"

# Check credentials in .env file
```

---

## Production Checklist

- [ ] DNS properly configured and resolving
- [ ] SSL certificates obtained and auto-renewal enabled
- [ ] All default passwords changed
- [ ] Google OAuth credentials configured
- [ ] Environment secrets generated and secure
- [ ] NAS devices configured in database
- [ ] VLAN rules configured and tested
- [ ] Firewall rules applied
- [ ] Database backups configured
- [ ] Monitoring and alerting set up
- [ ] Log rotation configured
- [ ] Documentation reviewed and customized
- [ ] Test authentication flow end-to-end
- [ ] Load testing performed (if high-volume expected)
- [ ] Security audit completed

---

## Support

For issues and questions:
- Check [TROUBLESHOOTING.md](./TROUBLESHOOTING.md)
- Review logs in `logs/` directory
- Check GitHub issues
- Contact support team

---

## Next Steps

After successful deployment:
1. Review [OPERATIONS.md](./OPERATIONS.md) for day-to-day operations
2. Set up monitoring and alerting
3. Train administrators on dashboard usage
4. Plan for scaling if needed
5. Schedule regular security audits
