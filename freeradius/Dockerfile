FROM ubuntu:22.04

LABEL maintainer="FreeRADIUS Google SSO Dashboard"
LABEL description="FreeRADIUS 3.x with MySQL support and dynamic VLAN assignment"

# Prevent interactive prompts during installation
ENV DEBIAN_FRONTEND=noninteractive

# Install FreeRADIUS and required modules
RUN apt-get update && apt-get install -y \
    freeradius \
    freeradius-mysql \
    freeradius-utils \
    mysql-client \
    curl \
    vim \
    gettext-base \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Create directories
RUN mkdir -p /var/log/freeradius \
    && chown -R freerad:freerad /var/log/freeradius \
    && mkdir -p /etc/freeradius/3.0/custom

# Copy only our custom configuration files (keep system's radiusd.conf with $INCLUDE directives)
# Don't copy radiusd.conf - use system default which properly includes all directories
COPY config/clients.conf /etc/freeradius/3.0/clients.conf
COPY config/mods-available/sql /etc/freeradius/3.0/mods-available/sql
COPY config/sites-available/default /etc/freeradius/3.0/sites-available/default
COPY entrypoint.sh /entrypoint.sh

# Make entrypoint executable and set proper permissions
RUN chmod +x /entrypoint.sh \
    && chmod -R o-w /etc/freeradius/3.0 \
    && chmod 750 /etc/freeradius/3.0

# Expose RADIUS ports
EXPOSE 1812/udp 1813/udp

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD radtest test test localhost 0 testing123 || exit 1

# Set working directory
WORKDIR /etc/freeradius/3.0

# Start FreeRADIUS (entrypoint will handle user switching)
ENTRYPOINT ["/entrypoint.sh"]
CMD ["freeradius", "-f", "-X"]
