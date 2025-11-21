# Stage 1: Builder
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code
COPY . .

# Build the application
# -ldflags="-s -w" strips debug information, reducing binary size
# CGO_ENABLED=0 is important for creating a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/quantaid ./cmd/qid-server

# Stage 2: Runner
FROM alpine:latest

# Install CA certificates for making HTTPS requests, and timezone data
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app/

# Create a non-root user and group
RUN addgroup -S quantaid && adduser -S quantaid -G quantaid

# Copy the compiled binary from the builder stage
COPY --from=builder /app/quantaid .

# Copy configuration files (optional, can be mounted as volume)
COPY configs/server.yaml.example ./configs/server.yaml.example

# Set ownership of the app directory to the new user
RUN chown -R quantaid:quantaid /app

# Switch to the non-root user
USER quantaid

# Expose the application port
EXPOSE 8080

# The command to run the application
CMD ["./quantaid"]
