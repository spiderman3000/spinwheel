Looking at your spinwheel project, I'll provide a comprehensive architecture recommendation with IDL contract design and backend setup using Bazel.

## Recommended Architecture

### 1. **IDL Choice: Protocol Buffers (Protobuf)**

**Why Protobuf:**
- Strong type safety across FE & BE
- Efficient serialization
- Excellent Go and TypeScript support
- Can use both gRPC (for performance) and REST (via gRPC-Gateway)
- Future-proof for microservices

## Project Structure

```
spinwheel/
├── WORKSPACE.bazel
├── .bazelrc
├── .bazelversion
├── frontend/                    # Existing React app
│   ├── BUILD.bazel
│   └── ...
├── backend/                     # New Go backend
│   ├── BUILD.bazel
│   ├── cmd/
│   │   └── server/
│   │       ├── BUILD.bazel
│   │       └── main.go
│   ├── internal/
│   │   ├── handler/
│   │   │   ├── BUILD.bazel
│   │   │   └── wheel.go
│   │   ├── service/
│   │   │   ├── BUILD.bazel
│   │   │   └── wheel.go
│   │   └── repository/
│   │       ├── BUILD.bazel
│   │       └── wheel.go
│   ├── pkg/
│   │   └── models/
│   │       ├── BUILD.bazel
│   │       └── wheel.go
│   └── go.mod
├── idl/                         # IDL contracts
│   ├── BUILD.bazel
│   ├── proto/
│   │   └── spinwheel/
│   │       └── v1/
│   │           ├── BUILD.bazel
│   │           ├── wheel.proto
│   │           └── service.proto
│   └── openapi/                 # Optional REST API spec
│       └── spinwheel.yaml
└── tools/
    └── BUILD.bazel
```

## IDL Contract Design## Summary

I've designed a complete backend architecture for your spinwheel project. Here's what I recommend:

### **IDL Choice: Protocol Buffers (Protobuf) with gRPC**

**Why this choice:**
- ✅ Strong type safety between Frontend & Backend
- ✅ Auto-generated code for Go and TypeScript
- ✅ Both gRPC (performance) and REST (via grpc-gateway) support
- ✅ Future-proof for microservices
- ✅ Industry standard with excellent tooling

### **Complete File Structure:**

```
spinwheel/
├── WORKSPACE.bazel              # Bazel workspace config
├── .bazelrc                     # Bazel settings
├── .bazelversion                # Bazel version pin
├── BUILD.bazel                  # Root build file with Gazelle
├── deps.bzl                     # Go dependencies (auto-generated)
│
├── frontend/                    # Your existing React app
│   ├── BUILD.bazel
│   └── ... (existing files)
│
├── idl/                         # **NEW: IDL contracts**
│   ├── BUILD.bazel
│   └── proto/
│       └── spinwheel/
│           └── v1/
│               ├── BUILD.bazel
│               ├── wheel.proto      # Data models
│               └── service.proto    # API definitions
│
├── backend/                     # **NEW: Go backend**
│   ├── BUILD.bazel
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   └── server/
│   │       ├── BUILD.bazel
│   │       └── main.go          # Entry point
│   ├── internal/
│   │   ├── handler/
│   │   │   ├── BUILD.bazel
│   │   │   └── wheel.go         # gRPC handlers
│   │   ├── service/
│   │   │   ├── BUILD.bazel
│   │   │   └── wheel.go         # Business logic
│   │   └── repository/
│   │       ├── BUILD.bazel
│   │       └── wheel.go         # Data access
│   └── pkg/
│       └── models/
│           └── BUILD.bazel
│
└── tools/
    └── BUILD.bazel
```

### **Key Features Implemented:**

1. **Dual Protocol Support:**
   - gRPC on port `:50051` (efficient binary protocol)
   - REST on port `:8080` (via grpc-gateway)

2. **Clean Architecture:**
   - Handler → Service → Repository layers
   - In-memory storage (easy to swap with PostgreSQL/MongoDB)

3. **Type-Safe Contracts:**
   - Proto definitions ensure FE/BE compatibility
   - Auto-generated client code

4. **Production Ready:**
   - CORS support
   - Graceful shutdown
   - Error handling
   - Pagination support

### **Next Steps:**

1. **Set up the directories** as shown above
2. **Copy the proto files** I created into `idl/proto/spinwheel/v1/`
3. **Follow the setup guide** I provided to initialize Bazel
4. **Build and run** the backend
5. **Update your React app** to call the REST API

### **Quick Start Commands:**

```bash
# 1. Install Bazel
brew install bazel  # or use bazelisk

# 2. Create directories
mkdir -p idl/proto/spinwheel/v1
mkdir -p backend/{cmd/server,internal/{handler,service,repository}}

# 3. Copy proto files and code

# 4. Initialize Go module
cd backend && go mod init github.com/yourusername/spinwheel

# 5. Build everything
bazel build //...

# 6. Run server
bazel run //backend/cmd/server:server
```

The artifacts I created include:
1. **Proto definitions** (wheel.proto & service.proto)
2. **Bazel configuration** (WORKSPACE.bazel)
3. **Complete Go backend** (main.go, repository, service, handler)
4. **Setup guide** with step-by-step instructions
5. **API documentation** with examples

This setup gives you a scalable, maintainable architecture that's ready for production!