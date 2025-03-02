FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY main.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o kbak .

# Use a minimal Alpine image for the final stage
FROM alpine:3.19

WORKDIR /

# Install CA certificates for HTTPS connections to Kubernetes API
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/kbak /kbak

# Create a directory for backup output
RUN mkdir -p /backups

# Use the backup directory as the default output
ENTRYPOINT ["/kbak", "--output", "/backups"]