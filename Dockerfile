# Multi-stage Dockerfile for GymRat

# Stage 1: Build binary
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy dependency definitions
COPY go.mod go.sum ./
RUN go mod download

# Copy source files
COPY . .

# Build lightweight static Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o gymrat main.go

# Stage 2: Minimal runtime image
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for secure HTTPs if needed
RUN apk --no-cache add ca-certificates tzdata

# Copy binary and static web assets from builder
COPY --from=builder /app/gymrat /app/gymrat
COPY --from=builder /app/web /app/web

# Create persistent data volume directory
RUN mkdir -p /app/data/sessions

# Default Environment Variables
ENV SERVER_MODE=true \
    PORT=8080 \
    SESSIONS_DIR=/app/data/sessions \
    WEB_DIR=/app/web

# Expose server port
EXPOSE 8080

# Volume for data persistence across container restarts
VOLUME ["/app/data"]

# Run GymRat in server mode
ENTRYPOINT ["/app/gymrat"]
CMD ["-server"]
