#!/bin/sh

# Check if PHP-FPM is running
if ! pgrep -x php-fpm > /dev/null; then
    echo "PHP-FPM is not running"
    exit 1
fi

# Check if Nginx is running
if ! pgrep -x nginx > /dev/null; then
    echo "Nginx is not running"
    exit 1
fi

# Check if web server responds
if ! wget --quiet --tries=1 --spider http://localhost/health.php 2>/dev/null; then
    echo "Web server not responding"
    exit 1
fi

echo "Health check passed"
exit 0
