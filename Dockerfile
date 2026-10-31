# Use the official Go image to build the Go program. --platform=$BUILDPLATFORM
# pins this stage to the host architecture even when cross-building for a
# different TARGETARCH (e.g. arm64 on an amd64 runner) - Go cross-compiles
# natively via GOARCH, so the build itself never needs emulation.
FROM --platform=$BUILDPLATFORM golang:1.26 AS builder

# Version info baked into the binary via -ldflags below. APP_VERSION is the
# MAJOR.MINOR from the repo-root VERSION file; BUILD_NUMBER is the total git
# commit count at build time (so it advances by exactly 1 per commit);
# COMMIT_SHA is the short git SHA. All three are computed by the caller
# (see .github/workflows/google-cloudrun-docker.yml) since this build
# context doesn't include the .git directory.
ARG APP_VERSION=dev
ARG BUILD_NUMBER=0
ARG COMMIT_SHA=unknown

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY src/ .

# Build the Go program with static link for smaller size and no libc dependencies.
# TARGETOS/TARGETARCH are populated automatically by buildx from --platform;
# on a plain `docker build` (no --platform) they default to the host's own.
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -installsuffix cgo \
    -ldflags "-X main.version=${APP_VERSION} -X main.buildNumber=${BUILD_NUMBER} -X main.commitSHA=${COMMIT_SHA}" \
    -o sbcwaste .

# Use distroless static image — no shell, no perl, no glibc, ca-certs included.
# The :nonroot tag runs as UID 65532 instead of root.
FROM gcr.io/distroless/static-debian13:nonroot

# Set working directory
WORKDIR /app

# Copy the compiled Go program from the builder stage
COPY --from=builder /app/sbcwaste .

# Copy the static assets
COPY static ./static

# Set the environment variable for the port. Cloud Run will set this value.
ENV PORT 8080

# Expose the port the app runs on
EXPOSE 8080

# Command to run the binary
CMD ["./sbcwaste"]
