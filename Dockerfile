# Use the official Go image to build the Go program
FROM golang:1.24.11 AS builder

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
FROM debian:trixie-slim

# Update the base image to include the latest security patches and CA certificates
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y ca-certificates && \
    update-ca-certificates

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