#!/bin/bash
set -e

# Set default values for environment variables
: ${MYSQL_HOST:=mysql}
: ${MYSQL_PORT:=3306}
: ${MYSQL_DATABASE:=radius}
: ${MYSQL_USER:=radius}
: ${MYSQL_PASSWORD:=radiuspassword}
: ${RADIUS_SECRET:=testing123}
: ${RADIUS_DEBUG_LEVEL:=-X}

echo "==================================="
echo "FreeRADIUS Startup Configuration"
echo "==================================="
echo "MySQL Host: ${MYSQL_HOST}"
echo "MySQL Port: ${MYSQL_PORT}"
echo "MySQL Database: ${MYSQL_DATABASE}"
echo "MySQL User: ${MYSQL_USER}"
echo "==================================="

# Wait for MySQL to be ready
echo "Waiting for MySQL to be ready..."
MAX_RETRIES=30
RETRY_COUNT=0

until mysql -h"${MYSQL_HOST}" -P"${MYSQL_PORT}" -u"${MYSQL_USER}" -p"${MYSQL_PASSWORD}" -e "SELECT 1" > /dev/null 2>&1; do
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
        echo "ERROR: MySQL did not become ready in time"
        exit 1
    fi
    echo "MySQL is unavailable - attempt $RETRY_COUNT/$MAX_RETRIES"
    sleep 2
done

echo "MySQL is up - proceeding with FreeRADIUS configuration"

# Function to replace ONLY our environment variables in files
# This prevents envsubst from corrupting FreeRADIUS variables like ${.:name} or %{User-Name}
replace_env_vars_sql() {
    local file=$1
    echo "Processing SQL configuration file: $file"

    if [ -f "$file" ]; then
        # Only substitute our specific environment variables, leaving FreeRADIUS variables intact
        envsubst '${MYSQL_HOST} ${MYSQL_PORT} ${MYSQL_DATABASE} ${MYSQL_USER} ${MYSQL_PASSWORD}' < "$file" > "${file}.tmp"
        if mv "${file}.tmp" "$file" 2>/dev/null; then
            echo "✓ Updated: $file"
        else
            rm -f "${file}.tmp"
            echo "⚠ Failed to update: $file"
        fi
    else
        echo "✗ File not found: $file"
    fi
}

replace_env_vars_clients() {
    local file=$1
    echo "Processing clients configuration file: $file"

    if [ -f "$file" ]; then
        # Only substitute RADIUS_SECRET
        envsubst '${RADIUS_SECRET}' < "$file" > "${file}.tmp"
        if mv "${file}.tmp" "$file" 2>/dev/null; then
            echo "✓ Updated: $file"
        else
            rm -f "${file}.tmp"
            echo "⚠ Failed to update: $file"
        fi
    else
        echo "✗ File not found: $file"
    fi
}

# Replace environment variables in all configuration files
echo "Substituting environment variables in configuration files..."

# SQL module configuration
if [ -f /etc/freeradius/3.0/mods-available/sql ]; then
    export MYSQL_HOST MYSQL_PORT MYSQL_DATABASE MYSQL_USER MYSQL_PASSWORD
    replace_env_vars_sql /etc/freeradius/3.0/mods-available/sql
fi

# Clients configuration
if [ -f /etc/freeradius/3.0/clients.conf ]; then
    export RADIUS_SECRET
    replace_env_vars_clients /etc/freeradius/3.0/clients.conf
fi

# Radiusd configuration (skip if using system default)
# if [ -f /etc/freeradius/3.0/radiusd.conf ]; then
#     replace_env_vars /etc/freeradius/3.0/radiusd.conf
# fi

# Sites configuration - DON'T process with envsubst as it corrupts FreeRADIUS variables like %{User-Name}
# for site_file in /etc/freeradius/3.0/sites-enabled/*; do
#     if [ -f "$site_file" ]; then
#         replace_env_vars "$site_file"
#     fi
# done

# Enable SQL module if not already enabled
if [ ! -L /etc/freeradius/3.0/mods-enabled/sql ]; then
    echo "Enabling SQL module..."
    mkdir -p /etc/freeradius/3.0/mods-enabled
    ln -sf ../mods-available/sql /etc/freeradius/3.0/mods-enabled/sql
fi

# Enable default site if not already enabled
if [ ! -L /etc/freeradius/3.0/sites-enabled/default ]; then
    echo "Enabling default site..."
    mkdir -p /etc/freeradius/3.0/sites-enabled
    ln -sf ../sites-available/default /etc/freeradius/3.0/sites-enabled/default
fi

# Enable inner-tunnel site if not already enabled
if [ ! -L /etc/freeradius/3.0/sites-enabled/inner-tunnel ]; then
    echo "Enabling inner-tunnel site..."
    ln -sf ../sites-available/inner-tunnel /etc/freeradius/3.0/sites-enabled/inner-tunnel
fi

# Enable required modules
echo "Enabling required modules..."
mkdir -p /etc/freeradius/3.0/mods-enabled
for mod in pap chap mschap eap preprocess filter files detail attr_filter expiration logintime; do
    if [ ! -L /etc/freeradius/3.0/mods-enabled/$mod ] && [ -f /etc/freeradius/3.0/mods-available/$mod ]; then
        ln -sf ../mods-available/$mod /etc/freeradius/3.0/mods-enabled/$mod
        echo "  Enabled module: $mod"
    fi
done

# Set proper permissions (ignore errors due to Docker volume mounts)
chmod -R o-w /etc/freeradius/3.0 2>/dev/null || true
chmod 750 /etc/freeradius/3.0 2>/dev/null || true
chown -R freerad:freerad /etc/freeradius/3.0 2>/dev/null || true
chown -R freerad:freerad /var/log/freeradius 2>/dev/null || true

# Test FreeRADIUS configuration
echo "Testing FreeRADIUS configuration..."
echo "Skipping config test - will start server directly"

echo "==================================="
echo "Starting FreeRADIUS Server..."
echo "Debug Level: ${RADIUS_DEBUG_LEVEL}"
echo "==================================="

# Execute the main command
exec "$@"
