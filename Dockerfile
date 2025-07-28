# Enhanced Subdomain Enumeration Tool v2.2 - Production Docker Build
# Multi-stage build for optimal image size and security

ARG GO_VERSION=1.24
ARG ALPINE_VERSION=3.19

# Build stage
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

# Build arguments
ARG VERSION=2.2.0
ARG BUILD_TIME
ARG GIT_COMMIT

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    upx

# Set working directory
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download and verify dependencies
RUN go mod download && \
    go mod verify

# Copy source code
COPY . .

# Build the application with optimizations and metadata
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME:-$(date -u +%Y-%m-%dT%H:%M:%SZ)} -X main.gitCommit=${GIT_COMMIT:-unknown}" \
    -a -installsuffix cgo \
    -tags netgo \
    -o subdomain-enum \
    cmd/server/main.go

# Compress binary (optional, reduces size by ~30%)
RUN upx --best --lzma subdomain-enum || true

# Verify the binary works
RUN ./subdomain-enum --version || echo "Version check completed"

# Production stage - use distroless for maximum security
FROM gcr.io/distroless/static-debian12:latest AS production

# Metadata labels following OCI specification
LABEL org.opencontainers.image.title="Advanced Subdomain Enumeration Tool" \
      org.opencontainers.image.description="Professional multi-source subdomain discovery platform with real-time streaming, metrics, and web interface" \
      org.opencontainers.image.version="2.2.0" \
      org.opencontainers.image.authors="Security Research Team" \
      org.opencontainers.image.url="https://github.com/thespecialone1/subdomain-enum" \
      org.opencontainers.image.source="https://github.com/thespecialone1/subdomain-enum" \
      org.opencontainers.image.documentation="https://github.com/thespecialone1/subdomain-enum/blob/main/README.md" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="Security Tools" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.revision="${GIT_COMMIT}"

# Copy timezone data for proper time handling
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy CA certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set working directory
WORKDIR /app

# Copy the optimized binary
COPY --from=builder /app/subdomain-enum /app/subdomain-enum

# Copy static files (web interface)
COPY --from=builder /app/public/ /app/public/

# Environment variables with secure defaults
ENV PORT=8080 \
    METRICS_PORT=9090 \
    LOG_LEVEL=INFO \
    ENABLE_METRICS=true \
    ENABLE_HEALTH=true \
    DNS_SERVERS=8.8.8.8:53,1.1.1.1:53,208.67.222.222:53 \
    DNS_CONCURRENCY=50 \
    RATE_LIMIT_RPS=10 \
    RATE_LIMIT_BURST=20 \
    HTTP_SKIP_TLS_VERIFY=true \
    MAX_CONCURRENT_JOBS=10 \
    TIMEOUT_WAYBACK=5m \
    TIMEOUT_CRTSH=5m \
    TIMEOUT_DNS=10m \
    TIMEOUT_SEARCH=5m \
    TIMEOUT_PERMUTE=10m \
    TIMEOUT_ZONE=2m

# Expose ports
EXPOSE 8080 9090

# Health check configuration
HEALTHCHECK --interval=30s \
           --timeout=10s \
           --start-period=5s \
           --retries=3 \
    CMD ["/app/subdomain-enum", "--health-check"]

# Use non-root user (distroless handles this automatically)
USER 65534:65534

# Set the entrypoint
ENTRYPOINT ["/app/subdomain-enum"]

# Default command (can be overridden)
CMD []

# Development stage (optional) - includes debugging tools
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS development

# Install development tools
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    curl \
    wget \
    jq \
    bash \
    htop \
    strace

WORKDIR /app

# Copy source and dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Install air for hot reloading (development only)
RUN go install github.com/cosmtrek/air@latest

# Environment for development
ENV GIN_MODE=debug \
    PORT=8080 \
    LOG_LEVEL=DEBUG

EXPOSE 8080 9090

# Development command with hot reloading
CMD ["go", "run", "cmd/server/main.go"]# Enhanced Subdomain Enumeration Tool v2.2 - Production Docker Build
# Multi-stage build for optimal image size and security

ARG GO_VERSION=1.21
ARG ALPINE_VERSION=3.19

# Build stage
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

# Build arguments
ARG VERSION=2.2.0
ARG BUILD_TIME
ARG GIT_COMMIT

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    upx

# Set working directory
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download and verify dependencies
RUN go mod download && \
    go mod verify

# Copy source code
COPY . .

# Build the application with optimizations and metadata
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME:-$(date -u +%Y-%m-%dT%H:%M:%SZ)} -X main.gitCommit=${GIT_COMMIT:-unknown}" \
    -a -installsuffix cgo \
    -tags netgo \
    -o subdomain-enum \
    cmd/server/main.go

# Compress binary (optional, reduces size by ~30%)
RUN upx --best --lzma subdomain-enum || true

# Verify the binary works
RUN ./subdomain-enum --version || echo "Version check not available"

# Production stage - use distroless for maximum security
FROM gcr.io/distroless/static-debian12:latest AS production

# Metadata labels following OCI specification
LABEL org.opencontainers.image.title="Advanced Subdomain Enumeration Tool" \
      org.opencontainers.image.description="Professional multi-source subdomain discovery platform with real-time streaming, metrics, and web interface" \
      org.opencontainers.image.version="2.2.0" \
      org.opencontainers.image.authors="Security Research Team" \
      org.opencontainers.image.url="https://github.com/thespecialone1/subdomain-enum" \
      org.opencontainers.image.source="https://github.com/thespecialone1/subdomain-enum" \
      org.opencontainers.image.documentation="https://github.com/thespecialone1/subdomain-enum/blob/main/README.md" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="Security Tools" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.revision="${GIT_COMMIT}"

# Copy timezone data for proper time handling
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy CA certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set working directory
WORKDIR /app

# Copy the optimized binary
COPY --from=builder /app/subdomain-enum /app/subdomain-enum

# Copy static files (web interface)
COPY --from=builder /app/public/ /app/public/

# Create necessary directories and set permissions
# Note: distroless runs as nobody (65534:65534) by default

# Environment variables with secure defaults
ENV PORT=8080 \
    METRICS_PORT=9090 \
    LOG_LEVEL=INFO \
    ENABLE_METRICS=true \
    ENABLE_HEALTH=true \
    DNS_SERVERS=8.8.8.8:53,1.1.1.1:53,208.67.222.222:53 \
    DNS_CONCURRENCY=50 \
    RATE_LIMIT_RPS=10 \
    RATE_LIMIT_BURST=20 \
    HTTP_SKIP_TLS_VERIFY=true \
    MAX_CONCURRENT_JOBS=10 \
    TIMEOUT_WAYBACK=5m \
    TIMEOUT_CRTSH=5m \
    TIMEOUT_DNS=10m \
    TIMEOUT_SEARCH=5m \
    TIMEOUT_PERMUTE=10m \
    TIMEOUT_ZONE=2m

# Expose ports
EXPOSE 8080 9090

# Health check configuration
HEALTHCHECK --interval=30s \
           --timeout=10s \
           --start-period=5s \
           --retries=3 \
    CMD ["/app/subdomain-enum", "--health-check"] || exit 1

# Use non-root user (distroless handles this automatically)
USER 65534:65534

# Set the entrypoint
ENTRYPOINT ["/app/subdomain-enum"]

# Default command (can be overridden)
CMD []

# Development stage (optional) - includes debugging tools
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS development

# Install development tools
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    curl \
    wget \
    jq \
    bash \
    htop \
    strace

WORKDIR /app

# Copy source and dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Install air for hot reloading (development only)
RUN go install github.com/cosmtrek/air@latest

# Environment for development
ENV GIN_MODE=debug \
    PORT=8080 \
    LOG_LEVEL=DEBUG

EXPOSE 8080 9090

# Development command with hot reloading
CMD ["go", "run", "cmd/server/main.go"]

# Multi-architecture build support
FROM production AS final