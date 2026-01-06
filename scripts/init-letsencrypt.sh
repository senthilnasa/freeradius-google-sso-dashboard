#!/bin/bash

# Let's Encrypt SSL Certificate Initialization Script
# This script helps set up SSL certificates for the first time

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}Let's Encrypt SSL Certificate Setup${NC}"
echo -e "${GREEN}============================================${NC}"
echo

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
else
    echo -e "${RED}Error: .env file not found!${NC}"
    echo "Please create .env file from .env.example"
    exit 1
fi

# Check required variables
if [ -z "$PORTAL_DOMAIN" ] || [ -z "$ADMIN_DOMAIN" ] || [ -z "$CERTBOT_EMAIL" ]; then
    echo -e "${RED}Error: Required environment variables not set!${NC}"
    echo "Please set: PORTAL_DOMAIN, ADMIN_DOMAIN, CERTBOT_EMAIL in .env"
    exit 1
fi

echo -e "${YELLOW}Configuration:${NC}"
echo "Portal Domain: $PORTAL_DOMAIN"
echo "Admin Domain: $ADMIN_DOMAIN"
echo "Email: $CERTBOT_EMAIL"
echo

# Check if domains resolve
echo -e "${YELLOW}Checking DNS resolution...${NC}"
if ! nslookup $PORTAL_DOMAIN > /dev/null 2>&1; then
    echo -e "${RED}Warning: $PORTAL_DOMAIN does not resolve!${NC}"
    echo "Make sure DNS is configured before continuing."
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

if ! nslookup $ADMIN_DOMAIN > /dev/null 2>&1; then
    echo -e "${RED}Warning: $ADMIN_DOMAIN does not resolve!${NC}"
    echo "Make sure DNS is configured before continuing."
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Staging or production
echo
echo -e "${YELLOW}Certificate Mode:${NC}"
echo "1) Staging (for testing - recommended first)"
echo "2) Production (rate limited)"
read -p "Select mode (1/2): " MODE

STAGING_FLAG=""
if [ "$MODE" == "1" ]; then
    STAGING_FLAG="--staging"
    echo -e "${YELLOW}Using staging environment (certificates won't be trusted)${NC}"
else
    echo -e "${YELLOW}Using production environment${NC}"
fi

# Create required directories
echo
echo -e "${YELLOW}Creating required directories...${NC}"
mkdir -p certbot/logs
mkdir -p nginx/ssl

# Create temporary self-signed certificates for initial nginx startup
echo
echo -e "${YELLOW}Creating temporary self-signed certificates...${NC}"
openssl req -x509 -nodes -newkey rsa:2048 -days 1 \
    -keyout "nginx/ssl/${PORTAL_DOMAIN}.key" \
    -out "nginx/ssl/${PORTAL_DOMAIN}.crt" \
    -subj "/CN=${PORTAL_DOMAIN}"

openssl req -x509 -nodes -newkey rsa:2048 -days 1 \
    -keyout "nginx/ssl/${ADMIN_DOMAIN}.key" \
    -out "nginx/ssl/${ADMIN_DOMAIN}.crt" \
    -subj "/CN=${ADMIN_DOMAIN}"

echo -e "${GREEN}✓ Temporary certificates created${NC}"

# Start nginx with temporary config
echo
echo -e "${YELLOW}Starting Nginx for ACME challenge...${NC}"
docker-compose up -d nginx

# Wait for nginx to start
echo "Waiting for Nginx to start..."
sleep 10

# Test if nginx is responding
if ! curl -k -s http://localhost > /dev/null; then
    echo -e "${RED}Error: Nginx is not responding!${NC}"
    docker-compose logs nginx
    exit 1
fi

echo -e "${GREEN}✓ Nginx is running${NC}"

# Request certificates for portal domain
echo
echo -e "${YELLOW}Requesting certificate for ${PORTAL_DOMAIN}...${NC}"
docker-compose run --rm certbot certonly \
    --webroot \
    --webroot-path=/var/www/certbot \
    --email $CERTBOT_EMAIL \
    --agree-tos \
    --no-eff-email \
    $STAGING_FLAG \
    -d $PORTAL_DOMAIN

# Request certificates for admin domain
echo
echo -e "${YELLOW}Requesting certificate for ${ADMIN_DOMAIN}...${NC}"
docker-compose run --rm certbot certonly \
    --webroot \
    --webroot-path=/var/www/certbot \
    --email $CERTBOT_EMAIL \
    --agree-tos \
    --no-eff-email \
    $STAGING_FLAG \
    -d $ADMIN_DOMAIN

# Update nginx configuration to use Let's Encrypt certificates
echo
echo -e "${YELLOW}Updating Nginx configuration...${NC}"

# Reload nginx
echo
echo -e "${YELLOW}Reloading Nginx...${NC}"
docker-compose exec nginx nginx -s reload

# Start certbot renewal service
echo
echo -e "${YELLOW}Starting certificate renewal service...${NC}"
docker-compose up -d certbot-renew

echo
echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}SSL Certificate Setup Complete!${NC}"
echo -e "${GREEN}============================================${NC}"
echo
echo -e "${YELLOW}Next steps:${NC}"
echo "1. If using staging certificates, test your setup"
echo "2. Once confirmed working, run this script again with production mode"
echo "3. The certificates will auto-renew every 12 hours"
echo "4. Check certificate status: docker-compose logs certbot-renew"
echo
echo -e "${GREEN}Your sites should now be accessible via HTTPS:${NC}"
echo "  Portal: https://$PORTAL_DOMAIN"
echo "  Admin:  https://$ADMIN_DOMAIN"
echo
