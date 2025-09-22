# --- Stage 1: Builder ---
# This stage compiles the Go application.
FROM golang:1.21-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code
COPY . .

# Build the qid-server binary
# -ldflags="-w -s" strips debugging information, reducing the binary size.
# CGO_ENABLED=0 disables CGO, creating a statically linked binary.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /qid-server ./cmd/qid-server

# Build the qid CLI tool (optional, but can be useful for debugging inside the container)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /qid ./cmd/qid


# --- Stage 2: Final Image ---
# This stage creates the final, minimal production image.
FROM alpine:latest

# It's a good practice to run as a non-root user
RUN addgroup -S quanta && adduser -S quanta -G quanta
USER quanta

# Set the working directory
WORKDIR /home/quanta

# Copy the compiled binary from the builder stage
COPY --from=builder /qid-server /usr/local/bin/qid-server
# COPY --from=builder /qid /usr/local/bin/qid

# Copy the configuration directory
# Note: In a real deployment, you would likely mount the configuration as a volume
# or use a config map instead of copying it into the image.
COPY --from=builder /app/configs /home/quanta/configs

# Expose the port the server listens on
EXPOSE 8080

# The command to run when the container starts
ENTRYPOINT ["qid-server"]

#Personal.AI order the ending
