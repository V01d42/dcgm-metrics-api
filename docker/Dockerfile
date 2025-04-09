# Build stage
FROM golang:1.24.2-alpine AS builder

# Install build tools and set up environment
WORKDIR /app
COPY go.mod go.sum ./
RUN apk add --no-cache git gcc musl-dev tzdata && \
    go mod download

# Build optimized binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o dcgm-metrics-api ./cmd/dcgm-metrics-api

# Runtime stage
FROM scratch

# Copy required certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Set up working directory and copy binary
WORKDIR /app
COPY --from=builder /app/dcgm-metrics-api .

EXPOSE 8080
CMD ["./dcgm-metrics-api"] 