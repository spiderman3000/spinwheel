# Spinwheel Project Skill

## Project Overview
A gamified wheel-of-fortune application with Preact frontend and Go gRPC backend.

## Architecture
```
spinwheel/
├── frontend/          # Preact + TypeScript + Vite
├── backend/           # Go gRPC service
│   ├── cmd/server/    # Server entry point
│   ├── internal/      # Handler, service, repository layers
│   └── pkg/           # Shared models and utilities
├── idl/               # Protocol Buffers definitions
└── BUILD.bazel        # Bazel build configuration
```

## Key Components

### Frontend (Preact + Vite)
- **Wheel.tsx**: Canvas-based spin wheel with custom rendering
- **Playground.tsx**: Main UI container with item management
- **AddItemForm.tsx**: Form to add new options to wheel
- **ItemList.tsx**: Displays and manages wheel items

### Backend (Go gRPC)
- **WheelService**: CRUD operations + spin functionality
- **Repository**: In-memory storage with interface pattern
- **Middleware**: User ID extraction from headers
- **Handler**: gRPC service implementation

## Development Commands

### Frontend
```bash
cd frontend
npm install
npm run dev          # Start dev server
npm run build        # Build for production
npm run lint         # Run ESLint
```

### Backend
```bash
make build-server    # Build backend
make run-server      # Start gRPC server (port 50051)
make test            # Run all tests
```

### API Testing
```bash
make list-services   # List available gRPC services
make create-wheel    # Create test wheel
make list-wheels     # List all wheels
```

## Code Patterns

### Frontend State Management
```typescript
// Items managed via useState in App.tsx
const [items, setItems] = useState<Item[]>(initialItems);

// Theme persisted to localStorage
localStorage.setItem('theme', 'dark');
```

### Canvas Rendering (Wheel.tsx)
- Uses `useCallback` for stable draw function
- Handles DPR for retina displays
- `easeOutQuint` for smooth spin animation
- Responsive via `ResizeObserver`

### Backend Layering
```
Handler → Service → Repository → In-Memory Storage
```

### gRPC Server Setup
- User ID extracted via middleware interceptor
- Reflection enabled in development only
- Configurable port via `PORT` environment variable

## Styling
- Vanilla CSS with dark mode support
- Theme toggled via `.dark` class on `<html>`
- Responsive layout with sidebar + wheel

## Build System
- **Bazel**: Primary build tool for backend
- **npm**: Frontend dependency management
- **Makefile**: Convenience commands for development

## Testing
```bash
bazel test //...           # Run all Bazel tests
cd frontend && npm run lint # Frontend linting
```

## Deployment
- Backend: Cloud Run compatible (uses `PORT` env var)
- Frontend: Static build output in `frontend/dist/`

## Common Tasks

### Adding a New Component
1. Create component in `frontend/src/components/`
2. Export from `index.ts`
3. Import in parent component

### Modifying gRPC API
1. Update proto files in `idl/proto/`
2. Run `bazel build //...` to regenerate
3. Update handler, service, and repository layers

### Adding Backend Endpoint
1. Add method to `WheelService` interface
2. Implement in `WheelService` struct
3. Add handler method
4. Update proto definition
