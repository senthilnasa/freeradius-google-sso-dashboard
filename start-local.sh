#!/bin/bash

# FreeRADIUS Google SSO Dashboard - Local Development Startup Script
# This script starts the application using local SSL certificates

set -e

echo "=================================================="
echo " FreeRADIUS Google SSO Dashboard - Local Start"
echo "=================================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}Warning: .env file not found!${NC}"
    echo "Creating .env from .env.example..."
    cp .env.example .env
    echo -e "${GREEN}Created .env file. Please edit it with your configuration.${NC}"
    echo ""
fi

# Check if SSL directory exists
if [ ! -d "./ssl" ]; then
    echo -e "${YELLOW}Creating SSL directory...${NC}"
    mkdir -p ./ssl
fi

# Check for SSL certificates
PORTAL_CERT="./ssl/radius.krea.edu.in.crt"
PORTAL_KEY="./ssl/radius.krea.edu.in.key"
ADMIN_CERT="./ssl/admin-radius.krea.edu.in.crt"
ADMIN_KEY="./ssl/admin-radius.krea.edu.in.key"

MISSING_CERTS=false

echo "Checking SSL certificates..."
echo ""

if [ ! -f "$PORTAL_CERT" ] || [ ! -f "$PORTAL_KEY" ]; then
    echo -e "${RED}✗ Missing: Portal SSL certificates${NC}"
    echo "  Expected: $PORTAL_CERT"
    echo "  Expected: $PORTAL_KEY"
    MISSING_CERTS=true
else
    echo -e "${GREEN}✓ Portal SSL certificates found${NC}"
fi

if [ ! -f "$ADMIN_CERT" ] || [ ! -f "$ADMIN_KEY" ]; then
    echo -e "${RED}✗ Missing: Admin Dashboard SSL certificates${NC}"
    echo "  Expected: $ADMIN_CERT"
    echo "  Expected: $ADMIN_KEY"
    MISSING_CERTS=true
else
    echo -e "${GREEN}✓ Admin Dashboard SSL certificates found${NC}"
fi

echo ""

if [ "$MISSING_CERTS" = true ]; then
    echo -e "${RED}ERROR: SSL certificates are missing!${NC}"
    echo ""
    echo "Please place your SSL certificates in the ./ssl directory:"
    echo "  - radius.krea.edu.in.crt"
    echo "  - radius.krea.edu.in.key"
    echo "  - admin-radius.krea.edu.in.crt"
    echo "  - admin-radius.krea.edu.in.key"
    echo ""
    echo "If you don't have SSL certificates, you can generate self-signed ones:"
    echo ""
    echo "# For Portal (radius.krea.edu.in)"
    echo "openssl req -x509 -nodes -days 365 -newkey rsa:2048 \\"
    echo "  -keyout ./ssl/radius.krea.edu.in.key \\"
    echo "  -out ./ssl/radius.krea.edu.in.crt \\"
    echo "  -subj \"/CN=radius.krea.edu.in\""
    echo ""
    echo "# For Admin Dashboard (admin-radius.krea.edu.in)"
    echo "openssl req -x509 -nodes -days 365 -newkey rsa:2048 \\"
    echo "  -keyout ./ssl/admin-radius.krea.edu.in.key \\"
    echo "  -out ./ssl/admin-radius.krea.edu.in.crt \\"
    echo "  -subj \"/CN=admin-radius.krea.edu.in\""
    echo ""
    exit 1
fi

# Create necessary directories
echo "Creating log directories..."
mkdir -p ./logs/freeradius
mkdir -p ./logs/captive-portal
mkdir -p ./logs/admin-dashboard
mkdir -p ./logs/nginx
mkdir -p ./logs/certbot
echo -e "${GREEN}✓ Log directories created${NC}"
echo ""

# Stop any running containers
echo "Stopping existing containers (if any)..."
docker-compose down 2>/dev/null || true
echo ""

# Build containers
echo "Building Docker containers..."
echo ""
docker-compose build --no-cache

echo ""
echo -e "${GREEN}✓ Build complete!${NC}"
echo ""

# Start services
echo "Starting services..."
echo ""

# Start MySQL first and wait for it to be healthy
echo "Starting MySQL database..."
docker-compose up -d mysql
echo "Waiting for MySQL to be ready..."
sleep 10

# Start FreeRADIUS
echo "Starting FreeRADIUS server..."
docker-compose up -d freeradius
sleep 5

# Start Captive Portal
echo "Starting Captive Portal..."
docker-compose up -d captive-portal
sleep 3

# Start Admin Dashboard
echo "Starting Admin Dashboard..."
docker-compose up -d admin-dashboard
sleep 3

# Start Nginx
echo "Starting Nginx reverse proxy..."
docker-compose up -d nginx
sleep 3

# Start Redis (optional)
echo "Starting Redis..."
docker-compose up -d redis

echo ""
echo -e "${GREEN}=================================================="
echo " Services Started Successfully!"
echo "==================================================${NC}"
echo ""

# Check service status
echo "Service Status:"
echo "==============="
docker-compose ps
echo ""

echo -e "${GREEN}Application URLs:${NC}"
echo "  Captive Portal: https://radius.krea.edu.in"
echo "  Admin Dashboard: https://admin-radius.krea.edu.in"
echo ""
echo -e "${YELLOW}Default Admin Credentials:${NC}"
echo "  Username: admin"
echo "  Password: admin123"
echo "  ${RED}⚠️  CHANGE THESE IMMEDIATELY!${NC}"
echo ""

echo "To view logs:"
echo "  docker-compose logs -f                  # All services"
echo "  docker-compose logs -f captive-portal   # Portal only"
echo "  docker-compose logs -f admin-dashboard  # Dashboard only"
echo "  docker-compose logs -f freeradius       # RADIUS only"
echo ""

echo "To stop services:"
echo "  docker-compose down"
echo ""

echo -e "${GREEN}Setup complete! 🚀${NC}"
