# PHP Admin Dashboard Dockerfile
FROM php:8.2-fpm-alpine

LABEL maintainer="FreeRADIUS Google SSO Dashboard"
LABEL description="Admin Dashboard with Advanced Reporting"

# Install system dependencies
RUN apk add --no-cache \
    nginx \
    supervisor \
    curl \
    zip \
    unzip \
    git \
    libzip-dev \
    libpng-dev \
    libjpeg-turbo-dev \
    freetype-dev \
    icu-dev \
    oniguruma-dev

# Install PHP extensions
RUN docker-php-ext-configure gd --with-freetype --with-jpeg \
    && docker-php-ext-install -j$(nproc) \
    pdo \
    pdo_mysql \
    mysqli \
    mbstring \
    zip \
    gd \
    intl \
    opcache \
    bcmath

# Install Composer
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

# Create application directory
WORKDIR /var/www/html

# Copy application files
COPY . .

# Install PHP dependencies
RUN composer install --no-dev --optimize-autoloader --no-interaction

# Set permissions
RUN chown -R www-data:www-data /var/www/html \
    && chmod -R 755 /var/www/html \
    && mkdir -p /var/www/html/logs \
    && chown -R www-data:www-data /var/www/html/logs \
    && mkdir -p /var/log/supervisor

# Copy Nginx configuration
COPY docker/nginx.conf /etc/nginx/nginx.conf
COPY docker/default.conf /etc/nginx/http.d/default.conf

# Copy Supervisor configuration
COPY docker/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Copy PHP configuration
COPY docker/php.ini /usr/local/etc/php/conf.d/custom.ini
COPY docker/php-fpm.conf /usr/local/etc/php-fpm.d/www.conf

# Expose port
EXPOSE 80

# Health check script
COPY docker/health-check.sh /usr/local/bin/health-check
RUN chmod +x /usr/local/bin/health-check

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD /usr/local/bin/health-check

# Start supervisor
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
