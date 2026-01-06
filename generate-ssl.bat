@echo off
REM Generate Self-Signed SSL Certificates for Local Testing

echo ==================================================
echo  Generating Self-Signed SSL Certificates
echo ==================================================
echo.

REM Check if OpenSSL is installed
where openssl >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: OpenSSL is not installed or not in PATH!
    echo.
    echo Please install OpenSSL:
    echo   - Download from: https://slproweb.com/products/Win32OpenSSL.html
    echo   - Or install via Chocolatey: choco install openssl
    echo.
    pause
    exit /b 1
)

REM Create SSL directory if it doesn't exist
if not exist ssl mkdir ssl

echo Generating SSL certificate for radius.krea.edu.in...
openssl req -x509 -nodes -days 365 -newkey rsa:2048 ^
  -keyout ssl\radius.krea.edu.in.key ^
  -out ssl\radius.krea.edu.in.crt ^
  -subj "/C=IN/ST=Andhra Pradesh/L=Sri City/O=Krea University/CN=radius.krea.edu.in"

if %ERRORLEVEL% EQU 0 (
    echo [OK] Generated radius.krea.edu.in certificate
) else (
    echo [ERROR] Failed to generate radius.krea.edu.in certificate
    pause
    exit /b 1
)

echo.
echo Generating SSL certificate for admin-radius.krea.edu.in...
openssl req -x509 -nodes -days 365 -newkey rsa:2048 ^
  -keyout ssl\admin-radius.krea.edu.in.key ^
  -out ssl\admin-radius.krea.edu.in.crt ^
  -subj "/C=IN/ST=Andhra Pradesh/L=Sri City/O=Krea University/CN=admin-radius.krea.edu.in"

if %ERRORLEVEL% EQU 0 (
    echo [OK] Generated admin-radius.krea.edu.in certificate
) else (
    echo [ERROR] Failed to generate admin-radius.krea.edu.in certificate
    pause
    exit /b 1
)

echo.
echo ==================================================
echo  SSL Certificates Generated Successfully!
echo ==================================================
echo.
echo Generated files:
echo   - ssl\radius.krea.edu.in.crt
echo   - ssl\radius.krea.edu.in.key
echo   - ssl\admin-radius.krea.edu.in.crt
echo   - ssl\admin-radius.krea.edu.in.key
echo.
echo NOTE: These are self-signed certificates for testing only.
echo Your browser will show a security warning, which you can bypass.
echo.
echo You can now run: start-local.bat
echo.
pause
