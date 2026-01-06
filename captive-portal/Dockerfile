# Multi-stage build for Go application
FROM golang:1.21-alpine AS builder

LABEL maintainer="FreeRADIUS Google SSO Dashboard"
LABEL description="Captive Portal with Google OAuth and Dynamic VLAN Assignment"

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o captive-portal \
    ./cmd/server/main.go

# Final stage - minimal runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata wget

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Create directories
RUN mkdir -p /app/logs /app/templates /app/static && \
    chown -R appuser:appuser /app

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/captive-portal .

# Copy templates and static files
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Change ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose application port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./captive-portal"]
