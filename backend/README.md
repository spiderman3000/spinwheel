# Spinwheel Backend

This document provides instructions on how to build, start, and test the backend service for the Spinwheel application.

## Overview

The backend is a Go-based application built with Bazel. It uses a clean, 3-layer architecture (Handler -> Service -> Repository) and supports both gRPC and REST protocols.

- **gRPC Server:** Port `50051`
- **REST Gateway:** Port `8080`

## Development Environment Setup

The backend is built and managed using [Bazel](https://bazel.build/).

### Install Bazel

```bash
sudo apt-get update && sudo apt-get install -y curl gnupg
curl https://bazel.build/bazel-release.pub.gpg | sudo apt-key add -
echo "deb [arch=amd64] https://storage.googleapis.com/bazel-apt stable jdk1.8" | sudo tee /etc/apt/sources.list.d/bazel.list
sudo apt-get update && sudo apt-get install -y bazel
```

## Building, Testing, and Running

**Build all targets:**
```bash
bazel build //...
```

**Run the backend server:**
```bash
bazel run //backend/cmd/server:server
```

**Run all tests:**
```bash
bazel test //...
```

## Testing the API

The backend provides a gRPC API. You can test it using `grpcurl`:

**List available services:**
```bash
grpcurl -plaintext localhost:50051 list
```

**Create a wheel:**
```bash
grpcurl -plaintext -H "x-user-id: test-user" \
  -d '{"name": "Test Wheel", "initial_items": ["Option 1", "Option 2", "Option 3"]}' \
  localhost:50051 spinwheel.v1.WheelService/CreateWheel
```

**Note:** The `x-user-id` header is required by the UserIDInterceptor middleware for local development.

## Code Structure

- `cmd/server/main.go`: Entry point for the backend server.
- `internal/handler`: gRPC handlers processing incoming requests.
- `internal/service`: Core business logic.
- `internal/repository`: Data access layer (in-memory store).
- `pkg/models`: Core data structures.
- `gen/proto`: Generated Go code from `.proto` files.

## Frontend (Reference)

The frontend is a lightweight **Preact** application located in the `/frontend` directory. It uses a custom Canvas-based wheel implementation for high performance and minimal dependencies.

## Security Considerations

- **Development:** Never commit `.env` files. The `X-User-Id` header is for local testing only.
- **Production:** Replace `X-User-Id` with JWT authentication, enable HTTPS, and use environment variables for sensitive data.
- **Rate Limiting:** Default is set to 100 req/min per user.
