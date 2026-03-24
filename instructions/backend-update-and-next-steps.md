# Spinwheel Backend: Code Analysis & Next Steps

## 📊 Current State Analysis

### ✅ What's Working

1. **Architecture Foundation**
   - Clean 3-layer architecture (Handler → Service → Repository)
   - gRPC + REST dual protocol support via grpc-gateway
   - Protocol Buffers for type-safe API contracts
   - Bazel build system properly configured

2. **Core Functionality**
   - CRUD operations for wheels (Create, Read, Update, Delete)
   - Spin functionality with random selection
   - In-memory storage implementation

3. **Build System**
   - Bazel workspace configured with Go rules
   - Proto compilation pipeline set up
   - Gazelle for automatic BUILD file generation

### ⚠️ Critical Issues to Fix

#### 1. **Import Path Bug in Proto Files**
**Location:** `idl/proto/spinwheel/v1/service.proto` (line 4)

```protobuf
// ❌ INCORRECT
import "proto/spinwheel/v1/wheel.proto";

// ✅ CORRECT
import "idl/proto/spinwheel/v1/wheel.proto";
```

**Impact:** Proto compilation will fail  
**Fix Priority:** 🔴 CRITICAL - Do this first

#### 2. **Typo in Error Handling**
**Location:** `backend/cmd/server/main.go` (line 73)

```go
// ❌ INCORRECT
err = spinwheelv1.RegisterWheelServiceHandler(context.Background(), gwmux, conn)
if err != nil {
    logFatalf("failed to register gateway: %v", err)  // Wrong function
}

// ✅ CORRECT
if err != nil {
    log.Fatalf("failed to register gateway: %v", err)
}

// Also remove the broken logFatalf function at the bottom
```

**Impact:** Gateway registration failures won't be properly reported  
**Fix Priority:** 🔴 HIGH

#### 3. **API Contract Mismatch**
The API documentation shows a more complete API than what's implemented:

**Documentation Shows:**
```json
{
  "wheel": {
    "user_id": "user123",
    "theme": "THEME_DARK",
    "items": [
      {
        "weight": 1.5,
        "color": "#FF5733"
      }
    ]
  }
}
```

**Current Implementation:**
```protobuf
message Wheel {
  string id = 1;
  string name = 2;
  repeated WheelItem items = 3;
  // Missing: user_id, theme, created_at, updated_at
}

message WheelItem {
  string id = 1;
  string text = 2;
  string color = 3;
  // Missing: weight, wheel_id, created_at
}
```

**Impact:** Frontend integration will fail  
**Fix Priority:** 🟡 MEDIUM

---

## 🎯 Immediate Next Steps (Week 1)

### Step 1: Fix Critical Bugs (Day 1)

```bash
# 1. Fix proto import path
sed -i '' 's|proto/spinwheel/v1/wheel.proto|idl/proto/spinwheel/v1/wheel.proto|' \
  idl/proto/spinwheel/v1/service.proto

# 2. Fix typo in main.go
# Edit backend/cmd/server/main.go manually:
# - Change logFatalf to log.Fatalf
# - Remove the logFatalf function definition

# 3. Rebuild
bazel clean
bazel build //...

# 4. Test
bazel run //backend/cmd/server:server
```

### Step 2: Align Proto Definitions with API Docs (Day 2)

**Update `idl/proto/spinwheel/v1/wheel.proto`:**

```protobuf
syntax = "proto3";

package spinwheel.v1;

import "google/protobuf/timestamp.proto";

option go_package = "github.com/yourusername/spinwheel/backend/gen/proto/spinwheel/v1;spinwheelv1";

enum Theme {
  THEME_UNSPECIFIED = 0;
  THEME_LIGHT = 1;
  THEME_DARK = 2;
}

message Wheel {
  string id = 1;
  string name = 2;
  string user_id = 3;
  repeated WheelItem items = 4;
  Theme theme = 5;
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp updated_at = 7;
}

message WheelItem {
  string id = 1;
  string option = 2;  // Renamed from 'text' to match docs
  string wheel_id = 3;
  double weight = 4;
  string color = 5;
  google.protobuf.Timestamp created_at = 6;
}
```

**Update `idl/proto/spinwheel/v1/service.proto`:**

```protobuf
syntax = "proto3";

package spinwheel.v1;

import "google/api/annotations.proto";
import "idl/proto/spinwheel/v1/wheel.proto";

option go_package = "github.com/yourusername/spinwheel/backend/gen/proto/spinwheel/v1;spinwheelv1";

service WheelService {
  rpc CreateWheel(CreateWheelRequest) returns (CreateWheelResponse) {
    option (google.api.http) = {
      post: "/v1/wheels"
      body: "*"
    };
  }

  rpc GetWheel(GetWheelRequest) returns (GetWheelResponse) {
    option (google.api.http) = {
      get: "/v1/wheels/{id}"
    };
  }

  rpc ListWheels(ListWheelsRequest) returns (ListWheelsResponse) {
    option (google.api.http) = {
      get: "/v1/wheels"
    };
  }

  rpc UpdateWheel(UpdateWheelRequest) returns (UpdateWheelResponse) {
    option (google.api.http) = {
      patch: "/v1/wheels/{wheel.id}"
      body: "wheel"
    };
  }

  rpc DeleteWheel(DeleteWheelRequest) returns (DeleteWheelResponse) {
    option (google.api.http) = {
      delete: "/v1/wheels/{id}"
    };
  }

  rpc AddItem(AddItemRequest) returns (AddItemResponse) {
    option (google.api.http) = {
      post: "/v1/wheels/{wheel_id}/items"
      body: "*"
    };
  }

  rpc UpdateItem(UpdateItemRequest) returns (UpdateItemResponse) {
    option (google.api.http) = {
      patch: "/v1/wheels/{wheel_id}/items/{item.id}"
      body: "item"
    };
  }

  rpc DeleteItem(DeleteItemRequest) returns (DeleteItemResponse) {
    option (google.api.http) = {
      delete: "/v1/wheels/{wheel_id}/items/{item_id}"
    };
  }

  rpc SpinWheel(SpinWheelRequest) returns (SpinWheelResponse) {
    option (google.api.http) = {
      post: "/v1/wheels/{wheel_id}/spin"
      body: "*"
    };
  }

  rpc GetSpinHistory(GetSpinHistoryRequest) returns (GetSpinHistoryResponse) {
    option (google.api.http) = {
      get: "/v1/wheels/{wheel_id}/history"
    };
  }
}

// Request/Response messages
message CreateWheelRequest {
  string name = 1;
  Theme theme = 2;
  repeated string initial_items = 3;
}

message CreateWheelResponse {
  Wheel wheel = 1;
}

message GetWheelRequest {
  string id = 1;
}

message GetWheelResponse {
  Wheel wheel = 1;
}

message ListWheelsRequest {
  int32 page_size = 1;
  string page_token = 2;
}

message ListWheelsResponse {
  repeated Wheel wheels = 1;
  string next_page_token = 2;
  int32 total_count = 3;
}

message UpdateWheelRequest {
  Wheel wheel = 1;
}

message UpdateWheelResponse {
  Wheel wheel = 1;
}

message DeleteWheelRequest {
  string id = 1;
}

message DeleteWheelResponse {}

message AddItemRequest {
  string wheel_id = 1;
  string option = 2;
  string color = 3;
  double weight = 4;
}

message AddItemResponse {
  WheelItem item = 1;
}

message UpdateItemRequest {
  string wheel_id = 1;
  WheelItem item = 2;
}

message UpdateItemResponse {
  WheelItem item = 1;
}

message DeleteItemRequest {
  string wheel_id = 1;
  string item_id = 2;
}

message DeleteItemResponse {}

message SpinWheelRequest {
  string wheel_id = 1;
  int64 seed = 2;  // Optional: for deterministic testing
}

message SpinResult {
  string id = 1;
  string wheel_id = 2;
  WheelItem selected_item = 3;
  string user_id = 4;
  google.protobuf.Timestamp spun_at = 5;
}

message SpinWheelResponse {
  SpinResult result = 1;
}

message GetSpinHistoryRequest {
  string wheel_id = 1;
  int32 page_size = 2;
  string page_token = 3;
}

message GetSpinHistoryResponse {
  repeated SpinResult results = 1;
  string next_page_token = 2;
  int32 total_count = 3;
}
```

### Step 3: Add User Context via Interceptor (Day 3)

**Create `backend/internal/middleware/user.go`:**

```go
package middleware

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type userIDKey struct{}

// UserIDFromContext retrieves the user ID from context
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey{}).(string)
	return userID, ok
}

// UserIDInterceptor extracts X-User-Id from metadata
func UserIDInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		userIDs := md.Get("x-user-id")
		if len(userIDs) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing x-user-id header")
		}

		// Add user ID to context
		ctx = context.WithValue(ctx, userIDKey{}, userIDs[0])
		
		return handler(ctx, req)
	}
}
```

**Update `backend/cmd/server/main.go`:**

```go
import (
	// ... existing imports
	"github.com/yourusername/spinwheel/backend/internal/middleware"
)

func main() {
	// ... existing code

	// Create gRPC server with interceptors
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.UserIDInterceptor()),
	)

	// ... rest of the code
}
```

### Step 4: Add Proper Logging (Day 4)

**Install structured logging:**

```bash
cd backend
go get -u go.uber.org/zap
```

**Create `backend/pkg/logger/logger.go`:**

```go
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func Init(environment string) error {
	var config zap.Config

	if environment == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	var err error
	Log, err = config.Build()
	if err != nil {
		return err
	}

	return nil
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
```

**Update main.go to use logger:**

```go
import (
	"github.com/yourusername/spinwheel/backend/pkg/logger"
)

func main() {
	// Initialize logger
	if err := logger.Init("development"); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Log.Info("Starting Spinwheel server")

	// ... rest of the code

	logger.Log.Info("gRPC server listening", zap.String("port", "50051"))
	logger.Log.Info("HTTP server listening", zap.String("port", "8080"))
}
```

---

## 🚀 Medium-Term Goals (Weeks 2-4)

### Week 2: Database Integration

#### Option A: PostgreSQL (Recommended for Production)

**1. Add database migrations:**

```bash
# Install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create migrations
mkdir -p backend/migrations

# Create first migration
cat > backend/migrations/000001_init.up.sql << 'EOF'
CREATE TABLE wheels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    theme VARCHAR(50) DEFAULT 'THEME_LIGHT',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE wheel_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wheel_id UUID NOT NULL REFERENCES wheels(id) ON DELETE CASCADE,
    option TEXT NOT NULL,
    weight DECIMAL(10,2) DEFAULT 1.0,
    color VARCHAR(7),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT positive_weight CHECK (weight > 0)
);

CREATE TABLE spin_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wheel_id UUID NOT NULL REFERENCES wheels(id) ON DELETE CASCADE,
    selected_item_id UUID NOT NULL REFERENCES wheel_items(id),
    user_id VARCHAR(255) NOT NULL,
    spun_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wheels_user_id ON wheels(user_id);
CREATE INDEX idx_wheel_items_wheel_id ON wheel_items(wheel_id);
CREATE INDEX idx_spin_results_wheel_id ON spin_results(wheel_id);
CREATE INDEX idx_spin_results_user_id ON spin_results(user_id);
EOF
```

**2. Create PostgreSQL repository:**

```go
// backend/internal/repository/postgres.go
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/yourusername/spinwheel/backend/pkg/models"
)

type PostgresWheelRepository struct {
	db *sql.DB
}

func NewPostgresWheelRepository(connString string) (*PostgresWheelRepository, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresWheelRepository{db: db}, nil
}

func (r *PostgresWheelRepository) CreateWheel(ctx context.Context, wheel *models.Wheel) (*models.Wheel, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert wheel
	wheelID := uuid.New().String()
	now := time.Now()
	
	_, err = tx.ExecContext(ctx,
		`INSERT INTO wheels (id, name, user_id, theme, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		wheelID, wheel.Name, wheel.UserID, wheel.Theme, now, now,
	)
	if err != nil {
		return nil, err
	}

	// Insert items
	for i := range wheel.Items {
		itemID := uuid.New().String()
		_, err = tx.ExecContext(ctx,
			`INSERT INTO wheel_items (id, wheel_id, option, weight, color, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			itemID, wheelID, wheel.Items[i].Option,
			wheel.Items[i].Weight, wheel.Items[i].Color, now,
		)
		if err != nil {
			return nil, err
		}
		wheel.Items[i].ID = itemID
		wheel.Items[i].WheelID = wheelID
		wheel.Items[i].CreatedAt = now
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	wheel.ID = wheelID
	wheel.CreatedAt = now
	wheel.UpdatedAt = now

	return wheel, nil
}

// Implement other methods...
```

#### Option B: MongoDB (For Rapid Prototyping)

```go
// backend/internal/repository/mongo.go
package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/yourusername/spinwheel/backend/pkg/models"
)

type MongoWheelRepository struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewMongoWheelRepository(uri, dbName string) (*MongoWheelRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &MongoWheelRepository{
		client: client,
		db:     client.Database(dbName),
	}, nil
}

func (r *MongoWheelRepository) CreateWheel(ctx context.Context, wheel *models.Wheel) (*models.Wheel, error) {
	collection := r.db.Collection("wheels")
	
	wheel.ID = primitive.NewObjectID().Hex()
	wheel.CreatedAt = time.Now()
	wheel.UpdatedAt = time.Now()
	
	for i := range wheel.Items {
		wheel.Items[i].ID = primitive.NewObjectID().Hex()
		wheel.Items[i].WheelID = wheel.ID
		wheel.Items[i].CreatedAt = time.Now()
	}

	_, err := collection.InsertOne(ctx, wheel)
	if err != nil {
		return nil, err
	}

	return wheel, nil
}

// Implement other methods...
```

### Week 3: Testing

**Create `backend/internal/handler/wheel_test.go`:**

```go
package handler_test

import (
	"context"
	"testing"

	"github.com/yourusername/spinwheel/backend/internal/handler"
	"github.com/yourusername/spinwheel/backend/internal/repository"
	"github.com/yourusername/spinwheel/backend/internal/service"
	spinwheelv1 "github.com/yourusername/spinwheel/backend/gen/proto/spinwheel/v1"
)

func TestCreateWheel(t *testing.T) {
	// Setup
	repo := repository.NewInMemoryWheelRepository()
	svc := service.NewWheelService(repo)
	handler := handler.NewWheelHandler(svc)

	// Test
	req := &spinwheelv1.CreateWheelRequest{
		Name: "Test Wheel",
		InitialItems: []string{"Option 1", "Option 2", "Option 3"},
	}

	resp, err := handler.CreateWheel(context.Background(), req)
	
	// Assert
	if err != nil {
		t.Fatalf("CreateWheel failed: %v", err)
	}
	
	if resp.Wheel.Name != "Test Wheel" {
		t.Errorf("Expected name 'Test Wheel', got '%s'", resp.Wheel.Name)
	}
	
	if len(resp.Wheel.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(resp.Wheel.Items))
	}
}

func TestSpinWheel(t *testing.T) {
	// Setup
	repo := repository.NewInMemoryWheelRepository()
	svc := service.NewWheelService(repo)
	handler := handler.NewWheelHandler(svc)

	// Create a wheel first
	createResp, _ := handler.CreateWheel(context.Background(), &spinwheelv1.CreateWheelRequest{
		Name: "Test Wheel",
		InitialItems: []string{"A", "B", "C"},
	})

	// Test spin
	spinReq := &spinwheelv1.SpinWheelRequest{
		WheelId: createResp.Wheel.Id,
		Seed:    12345,  // Deterministic for testing
	}

	spinResp, err := handler.SpinWheel(context.Background(), spinReq)
	
	// Assert
	if err != nil {
		t.Fatalf("SpinWheel failed: %v", err)
	}
	
	if spinResp.Result == nil {
		t.Fatal("Expected spin result, got nil")
	}
	
	if spinResp.Result.SelectedItem == nil {
		t.Fatal("Expected selected item, got nil")
	}
}
```

**Run tests:**

```bash
bazel test //backend/...
```

### Week 4: Configuration & Environment Management

**Create `backend/pkg/config/config.go`:**

```go
package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Environment  string
	GRPCPort     string
	HTTPPort     string
	DatabaseType string
	DatabaseURL  string
	LogLevel     string
}

func Load() (*Config, error) {
	cfg := &Config{
		Environment:  getEnv("ENVIRONMENT", "development"),
		GRPCPort:     getEnv("GRPC_PORT", "50051"),
		HTTPPort:     getEnv("HTTP_PORT", "8080"),
		DatabaseType: getEnv("DB_TYPE", "memory"),
		DatabaseURL:  getEnv("DATABASE_URL", ""),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.DatabaseType == "postgres" && c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required when DB_TYPE is postgres")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
```

**Create `.env.example`:**

```bash
# Environment
ENVIRONMENT=development

# Server Ports
GRPC_PORT=50051
HTTP_PORT=8080

# Database
DB_TYPE=postgres
DATABASE_URL=postgres://user:password@localhost:5432/spinwheel?sslmode=disable

# Logging
LOG_LEVEL=debug
```

---

## 🎯 Long-Term Roadmap (Months 2-3)

### Authentication & Authorization

1. **JWT Token Implementation**
   - Replace X-User-Id with proper JWT
   - Add token validation middleware
   - Implement refresh token flow

2. **OAuth2 Integration**
   - Google Sign-In
   - GitHub OAuth

### Advanced Features

1. **Weighted Spinning Algorithm**
   - Implement proper weighted random selection
   - Add animation seed for frontend consistency

2. **Wheel Sharing**
   - Public/private wheels
   - Share links
   - Collaborative wheels

3. **Analytics**
   - Track spin history
   - Most common results
   - User statistics

4. **Rate Limiting**
   - Per-user limits
   - Per-wheel limits
   - Redis-based distributed rate limiting

### Infrastructure

1. **Docker & Kubernetes**
   - Multi-stage Docker builds
   - K8s manifests
   - Helm charts

2. **CI/CD Pipeline**
   - GitHub Actions
   - Automated testing
   - Deployment pipelines

3. **Observability**
   - OpenTelemetry integration
   - Prometheus metrics
   - Grafana dashboards
   - Distributed tracing

---

## 📝 Useful Commands Reference

### Development

```bash
# Build everything
bazel build //...

# Run server
bazel run //backend/cmd/server:server

# Run tests
bazel test //backend/...

# Run specific test
bazel test //backend/internal/handler:handler_test

# Clean build cache
bazel clean --expunge
```

### Database Migrations

```bash
# Create new migration
migrate create -ext sql -dir backend/migrations -seq add_users_table

# Run migrations
migrate -path backend/migrations -database $DATABASE_URL up

# Rollback last migration
migrate -path backend/migrations -database $DATABASE_URL down 1

# Check migration status
migrate -path backend/migrations -database $DATABASE_URL version
```

### Testing API with cURL

```bash
# Create wheel
WHEEL_ID=$(curl -s -X POST http://localhost:8080/v1/wheels \
  -H "Content-Type: application/json" \
  -H "X-User-Id: test-user" \
  -d '{
    "name": "Lunch Options",
    "theme": "THEME_LIGHT",
    "initial_items": ["Pizza", "Sushi", "Tacos", "Burgers"]
  }' | jq -r '.wheel.id')

echo "Created wheel: $WHEEL_ID"

# Get wheel
curl http://localhost:8080/v1/wheels/$WHEEL_ID \
  -H "X-User-Id: test-user" | jq

# Spin wheel
curl -X POST http://localhost:8080/v1/wheels/$WHEEL_ID/spin \
  -H "X-User-Id: test-user" | jq '.result.selected_item.option'

# List wheels
curl http://localhost:8080/v1/wheels?page_size=10 \
  -H "X-User-Id: test-user" | jq

# Delete wheel
curl -X DELETE http://localhost:8080/v1/wheels/$WHEEL_ID \
  -H "X-User-Id: test-user"
```

---

## 🐛 Known Issues & Tech Debt

### Priority 1 (Fix ASAP)
- [ ] Proto import path bug
- [ ] logFatalf typo
- [ ] Missing error handling in repository layer
- [ ] No graceful shutdown implementation

### Priority 2 (Next Sprint)
- [ ] No input validation
- [ ] No pagination in ListWheels
- [ ] Random seed not used deterministically
- [ ] No transaction support in in-memory repo

### Priority 3 (Future)
- [ ] No caching layer
- [ ] No connection pooling
- [ ] Hardcoded CORS settings
- [ ] No health check endpoint

---

## 📚 Additional Resources

- [gRPC Go Tutorial](https://grpc.io/docs/languages/go/quickstart/)
- [gRPC-Gateway Documentation](https://grpc-ecosystem.github.io/grpc-gateway/)
- [Bazel Go Rules](https://github.com/bazelbuild/rules_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Protocol Buffers Guide](https://developers.google.com/protocol-buffers/docs/proto3)

---

## 💡 Pro Tips

1. **Use `buf` for proto management**: Consider migrating from raw protoc to [buf.build](https://buf.build/) for better proto linting and breaking change detection.

2. **Implement health checks**: Add gRPC health check protocol for better Kubernetes integration.

3. **Use pgx instead of database/sql**: For PostgreSQL, [pgx](https://github.com/jackc/pgx) offers better performance.

4. **Add request ID tracking**: Propagate request IDs through the call chain for better debugging.

5. **Consider CQRS pattern**: As the app grows, separate read and write models for better scalability.

---

**Last Updated:** November 2025  
**Maintainer:** [Your Name]  
**Status:** 🚧 In Active Development