# Spinwheel

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://github.com/yourusername/spinwheel/workflows/CI/badge.svg)](https://github.com/yourusername/spinwheel/actions)

A gamified wheel-of-fortune experience with a lightweight Preact frontend and Go backend.

## Features
- **Ultra-lightweight:** Preact-powered frontend (~19kB bundle).
- **Custom Canvas Wheel:** High-performance, dependency-free spin wheel.
- **Minimalist Design:** Clean, modern UI with vanilla CSS and dark mode support.
- **gRPC + REST API backend:** Robust Go-based service.
- **Type-safe:** End-to-end type safety with TypeScript and Protocol Buffers.

## Prerequisites

- [Node.js](https://nodejs.org/) 18+ and npm
- [Go](https://go.dev/dl/) 1.21+
- [Docker](https://docs.docker.com/get-docker/) (recommended for running the backend)
- [grpcurl](https://github.com/fullstorydev/grpcurl) (optional, for API testing)

## Quick Start

### Option 1: Docker (Recommended)

```bash
# Build and run the backend
docker build -t spinwheel .
docker run -p 50051:8080 spinwheel
```

### Option 2: Local Development

```bash
# Backend
cd backend
go mod tidy
go run ./cmd/server

# Frontend (in a separate terminal)
cd frontend
npm install
npm run dev
```

### Option 3: Bazel (requires Bazel/Bazelisk)

```bash
bazel run //backend/cmd/server:server
```

## API Testing

Once the backend is running, test with `grpcurl`:

```bash
# List available services
grpcurl -plaintext localhost:50051 list

# Create a wheel
grpcurl -plaintext -H "x-user-id: test-user" \
  -d '{"name": "Test Wheel", "initial_items": ["Red", "Blue", "Green", "Yellow"]}' \
  localhost:50051 spinwheel.v1.WheelService/CreateWheel

# List all wheels
grpcurl -plaintext -H "x-user-id: test-user" \
  localhost:50051 spinwheel.v1.WheelService/ListWheels

# Spin a wheel
grpcurl -plaintext -H "x-user-id: test-user" \
  -d '{"wheel_id": "<wheel-id>"}' \
  localhost:50051 spinwheel.v1.WheelService/SpinWheel
```

## Project Structure

```
spinwheel/
├── frontend/          # Preact + TypeScript + Vite
├── backend/           # Go gRPC service
│   ├── cmd/server/    # Server entry point
│   ├── internal/      # Handler, service, repository layers
│   └── pkg/           # Shared models and utilities
├── idl/               # Protocol Buffers definitions
├── Dockerfile         # Container build
└── Makefile           # Development commands
```

## Documentation
- [Frontend README](frontend/README.md)
- [Backend README](backend/README.md)
- [API Documentation](API%20Documentation%20-%20REST%20Endpoints.md)

## Contributing
Pull requests are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) first.

## Security
See [SECURITY.md](SECURITY.md) for reporting vulnerabilities.

## License
[MIT](LICENSE)
