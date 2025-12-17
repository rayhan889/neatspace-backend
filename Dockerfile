# ============================================
# Stage 1: Dependencies Cache Layer
# ============================================
FROM golang:1.25.1-alpine AS deps

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files for dependency caching
COPY go.mod go.sum ./

# Download dependencies with better caching
RUN go mod download && go mod verify

# ============================================
# Stage 2: Build Layer
# ============================================
FROM deps AS builder

# Copy source code
COPY . .

# Install swag CLI for generating API documentation
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger documentation
RUN swag init -g cmd/main.go -o docs --parseDependency --parseInternal

# Build the application with optimizations
# -s: omit symbol table and debug info
# -w: omit DWARF symbol table
# -trimpath: remove file system paths from binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-s -w -X main.version=${VERSION:-dev} -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -trimpath \
    -o neatspace \
    ./cmd/main.go

# Verify the binary was built successfully
RUN test -f neatspace && chmod +x neatspace

# ============================================
# Stage 3: Runtime Layer (Production)
# ============================================
FROM alpine:latest AS production

# Install runtime dependencies
# - ca-certificates: for HTTPS requests
# - tzdata: for timezone support
# - wget: for healthcheck
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    wget \
    && rm -rf /var/cache/apk/*

WORKDIR /app

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy binary from builder
COPY --from=builder --chown=appuser:appgroup /build/neatspace .

# Copy migrations, templates, and static files if they exist
COPY --from=builder --chown=appuser:appgroup /build/migrations ./migrations
COPY --from=builder --chown=appuser:appgroup /build/templates ./templates
COPY --from=builder --chown=appuser:appgroup /build/web ./web

# Create necessary directories with proper permissions
RUN mkdir -p storage logs && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose the application port (default 8080, can be overridden)
EXPOSE 8080

# Add health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

# Set the entrypoint
ENTRYPOINT ["./neatspace"]

# Default command
CMD ["serve"]