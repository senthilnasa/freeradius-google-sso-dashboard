@echo off
REM FreeRADIUS Google SSO Dashboard - Local Development Startup Script (Windows)
REM This script starts the application using local SSL certificates

echo ==================================================
echo  FreeRADIUS Google SSO Dashboard - Local Start
echo ==================================================
echo.

REM Check if .env file exists
if not exist .env (
    echo Warning: .env file not found!
    echo Creating .env from .env.example...
    copy .env.example .env
    echo Created .env file. Please edit it with your configuration.
    echo.
)

REM Check if SSL directory exists
if not exist ssl mkdir ssl

REM Check for SSL certificates
set PORTAL_CERT=ssl\radius.krea.edu.in.crt
set PORTAL_KEY=ssl\radius.krea.edu.in.key
set ADMIN_CERT=ssl\admin-radius.krea.edu.in.crt
set ADMIN_KEY=ssl\admin-radius.krea.edu.in.key

set MISSING_CERTS=false

echo Checking SSL certificates...
echo.

if not exist "%PORTAL_CERT%" (
    echo [ERROR] Missing: %PORTAL_CERT%
    set MISSING_CERTS=true
) else if not exist "%PORTAL_KEY%" (
    echo [ERROR] Missing: %PORTAL_KEY%
    set MISSING_CERTS=true
) else (
    echo [OK] Portal SSL certificates found
)

if not exist "%ADMIN_CERT%" (
    echo [ERROR] Missing: %ADMIN_CERT%
    set MISSING_CERTS=true
) else if not exist "%ADMIN_KEY%" (
    echo [ERROR] Missing: %ADMIN_KEY%
    set MISSING_CERTS=true
) else (
    echo [OK] Admin Dashboard SSL certificates found
)

echo.

if "%MISSING_CERTS%"=="true" (
    echo ERROR: SSL certificates are missing!
    echo.
    echo Please place your SSL certificates in the ./ssl directory:
    echo   - radius.krea.edu.in.crt
    echo   - radius.krea.edu.in.key
    echo   - admin-radius.krea.edu.in.crt
    echo   - admin-radius.krea.edu.in.key
    echo.
    echo If you need to generate self-signed certificates, use:
    echo   start-generate-ssl.bat
    echo.
    pause
    exit /b 1
)

REM Create necessary directories
echo Creating log directories...
if not exist logs\freeradius mkdir logs\freeradius
if not exist logs\captive-portal mkdir logs\captive-portal
if not exist logs\admin-dashboard mkdir logs\admin-dashboard
if not exist logs\nginx mkdir logs\nginx
if not exist logs\certbot mkdir logs\certbot
echo [OK] Log directories created
echo.

REM Stop any running containers
echo Stopping existing containers (if any)...
docker-compose down 2>nul
echo.

REM Build containers
echo Building Docker containers...
echo.
docker-compose build --no-cache

echo.
echo [OK] Build complete!
echo.

REM Start services
echo Starting services...
echo.

REM Start MySQL first and wait for it to be healthy
echo Starting MySQL database...
docker-compose up -d mysql
echo Waiting for MySQL to be ready...
timeout /t 10 /nobreak >nul

REM Start FreeRADIUS
echo Starting FreeRADIUS server...
docker-compose up -d freeradius
timeout /t 5 /nobreak >nul

REM Start Captive Portal
echo Starting Captive Portal...
docker-compose up -d captive-portal
timeout /t 3 /nobreak >nul

REM Start Admin Dashboard
echo Starting Admin Dashboard...
docker-compose up -d admin-dashboard
timeout /t 3 /nobreak >nul

REM Start Nginx
echo Starting Nginx reverse proxy...
docker-compose up -d nginx
timeout /t 3 /nobreak >nul

REM Start Redis (optional)
echo Starting Redis...
docker-compose up -d redis

echo.
echo ==================================================
echo  Services Started Successfully!
echo ==================================================
echo.

REM Check service status
echo Service Status:
echo ===============
docker-compose ps
echo.

echo Application URLs:
echo   Captive Portal: https://radius.krea.edu.in
echo   Admin Dashboard: https://admin-radius.krea.edu.in
echo.
echo Default Admin Credentials:
echo   Username: admin
echo   Password: admin123
echo   WARNING: CHANGE THESE IMMEDIATELY!
echo.

echo To view logs:
echo   docker-compose logs -f                  # All services
echo   docker-compose logs -f captive-portal   # Portal only
echo   docker-compose logs -f admin-dashboard  # Dashboard only
echo   docker-compose logs -f freeradius       # RADIUS only
echo.

echo To stop services:
echo   docker-compose down
echo.

echo Setup complete!
echo.
pause
