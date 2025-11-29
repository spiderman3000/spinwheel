# Spinwheel

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://github.com/yourusername/spinwheel/workflows/CI/badge.svg)](https://github.com/yourusername/spinwheel/actions)

A gamified wheel-of-fortune experience with React frontend and Go backend.

🎯 [Live Demo](https://yourusername.github.io/spinwheel/)

## Features
- Interactive spinning wheel
- Dark/light theme support
- Clean, modern UI with glassmorphism design
- gRPC + REST API backend
- Type-safe Protocol Buffers

## Quick Start

### Frontend
\`\`\`bash
cd frontend
npm install
npm run dev
\`\`\`

### Backend
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