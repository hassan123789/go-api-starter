# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies (git for version info, ca-certificates for HTTPS)
RUN apk add --no-cache git ca-certificates tzdata

# Create non-root user for the final stage
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with security flags
ARG VERSION=dev
ARG BUILD_TIME
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" \
    -o /server ./cmd/server

# Security scan stage (optional, for CI)
FROM golang:1.25-alpine AS security
RUN go install golang.org/x/vuln/cmd/govulncheck@latest
COPY --from=builder /app /app
WORKDIR /app
RUN govulncheck ./... || true

# Production stage - using distroless for minimal attack surface
FROM gcr.io/distroless/static-debian12:nonroot AS production

# Copy timezone data and CA certificates
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /server /server

# Use non-root user (distroless nonroot user has UID 65532)
USER nonroot:nonroot

# Expose port
EXPOSE 8080

# Health check
# Note: distroless doesn't have shell, so we can't use HEALTHCHECK here
# Use Kubernetes/Docker Compose health checks instead

# Run the application
ENTRYPOINT ["/server"]

# Alternative: Alpine-based production image (if you need shell access for debugging)
FROM alpine:3.23 AS production-alpine

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata && \
    rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

WORKDIR /app

# Copy the binary
COPY --from=builder /server .

# Set proper permissions
RUN chown -R appuser:appgroup /app

# Use non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./server"]
