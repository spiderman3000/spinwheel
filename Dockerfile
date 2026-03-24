# --- Stage 1: Build ---
FROM golang:1.22-bookworm AS builder

# Install dependencies: Node.js/npm (for Bazelisk), Java (for Bazel), protobuf compiler, and git
RUN apt-get update && apt-get install -y \
    nodejs npm \
    default-jre-headless \
    protobuf-compiler \
    git \
    && rm -rf /var/lib/apt/lists/*

# Install Bazelisk (Bazel wrapper)
RUN npm install -g @bazel/bazelisk

# Setup workspace
WORKDIR /src
COPY . .

# Generate go.sum if it doesn't exist (required by Bazel)
RUN cd backend && go mod tidy

# Build the backend server using Bazel
# This automatically handles the .proto generation
RUN bazelisk build //backend/cmd/server:server

# Move the binary to a known location to make copying easier
# The bazel output path is usually bazel-bin/backend/cmd/server/server_/server
RUN cp bazel-bin/backend/cmd/server/server_/server /app_binary

# --- Stage 2: Runtime ---
# Use a lightweight "distroless" image for security and size
FROM gcr.io/distroless/base-debian12

WORKDIR /

# Copy the binary from the builder stage
COPY --from=builder /app_binary /server

# Cloud Run injects the PORT env var, but we expose 8080 as documentation
ENV PORT=8080
EXPOSE 8080

# Run the binary
CMD ["/server"]