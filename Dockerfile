# Use the official Go image to build the Go program
FROM golang:1.24.13 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY src/ .

# Build the Go program with static link for smaller size and no libc dependencies
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sbcwaste .

# Use a slim base image
FROM debian:bookworm-slim

# Update the base image to include the latest security patches and CA certificates
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y --no-install-recommends ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy the compiled Go program from the builder stage
COPY --from=builder /app/sbcwaste /

# Copy the static assets
COPY static /static

# Set the environment variable for the port. Cloud Run will set this value.
ENV PORT 8080

# Expose the port the app runs on
EXPOSE 8080

# Run as a non-root user for security
USER nobody

# Health check hitting the /health endpoint
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:${PORT}/health || exit 1

# Command to run the binary
CMD ["/sbcwaste"]