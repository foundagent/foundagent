# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o foundagent ./cmd/foundagent

# Final stage
FROM alpine:latest

WORKDIR /root

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates git

# Copy binary from builder
COPY --from=builder /build/foundagent /usr/local/bin/fa

# Create symlink for full command name
RUN ln -s /usr/local/bin/fa /usr/local/bin/foundagent

# Set up non-root user
RUN addgroup -S foundagent && adduser -S foundagent -G foundagent
USER foundagent

# Set entrypoint
ENTRYPOINT ["fa"]
CMD ["--help"]
