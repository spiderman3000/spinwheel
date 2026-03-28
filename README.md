# Spinwheel

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://github.com/yourusername/spinwheel/workflows/CI/badge.svg)](https://github.com/yourusername/spinwheel/actions)

A gamified wheel-of-fortune experience with a lightweight Preact frontend and Go backend.

🎯 [Live Demo](https://yourusername.github.io/spinwheel/)

## Features
- **Ultra-lightweight:** Preact-powered frontend (~19kB bundle).
- **Custom Canvas Wheel:** High-performance, dependency-free spin wheel.
- **Minimalist Design:** Clean, modern UI with vanilla CSS and dark mode support.
- **gRPC + REST API backend:** Robust Go-based service.
- **Type-safe:** End-to-end type safety with TypeScript and Protocol Buffers.

## Quick Start

### Frontend
\`\`\`bash
cd frontend
npm install
npm run dev
\`\`\`

### Backend - Unimplemented
\`\`\`bash
bazel run //backend/cmd/server:server
\`\`\`

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
