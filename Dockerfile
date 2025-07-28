# Enhanced Subdomain Enumeration Tool - Multi-stage Docker build
FROM golang:1.24-alpine AS builder

# Install git and other build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=2.0.0 -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -a -installsuffix cgo \
    -o main cmd/server/main.go

# Runtime stage - use distroless for security
FROM gcr.io/distroless/static-debian11:latest

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy CA certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main ./main

# Copy static files (web interface)
COPY --from=builder /app/public ./public

# Create non-root user (distroless handles this automatically)
USER 65534:65534

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/app/main", "-health-check"]

# Set metadata labels
LABEL org.opencontainers.image.title="Advanced Subdomain Enumeration Tool"
LABEL org.opencontainers.image.description="Multi-source subdomain discovery with real-time streaming"
LABEL org.opencontainers.image.version="2.0.0"
LABEL org.opencontainers.image.authors="Security Research Team"
LABEL org.opencontainers.image.source="https://github.com/thespecialone1/subdomain-enum"

# Run the application
ENTRYPOINT ["/app/main"]