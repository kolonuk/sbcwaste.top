# Use the official Go image to build the Go program
FROM golang:1.24.3 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY src/ .

# Build the Go program with static link for smaller size and no libc dependencies
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sbcwaste .

# Use a distroless base image for security and a smaller footprint
FROM debian:bullseye-slim

# Install necessary dependencies for Chrome/Chromium
RUN apt-get update && apt-get install -y \
    ca-certificates \
    fonts-liberation \
    libappindicator3-1 \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libcups2 \
    libdbus-1-3 \
    libgdk-pixbuf2.0-0 \
    libnspr4 \
    libnss3 \
    libx11-xcb1 \
    lsb-release \
    wget \
    xdg-utils \
    chromium \
    chromium-driver \
    && rm -rf /var/lib/apt/lists/*

# Copy the compiled Go program from the builder stage
COPY --from=builder /app/sbcwaste /

# Copy the static assets
COPY static /static

# Set the environment variable for the port. Cloud Run will set this value.
ENV PORT 8080

# Expose the port the app runs on
EXPOSE 8080

# Command to run the binary
CMD ["/sbcwaste"]