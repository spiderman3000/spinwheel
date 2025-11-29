.PHONY: help build build-server test run-server run-tests list-services create-wheel get-wheel list-wheels clean

help:
	@echo "Spinwheel Development Commands"
	@echo "=============================="
	@echo ""
	@echo "Building:"
	@echo "  make build              Build all targets"
	@echo "  make build-server       Build backend server only"
	@echo ""
	@echo "Running:"
	@echo "  make run-server         Start the backend gRPC server (port 50051)"
	@echo ""
	@echo "Testing:"
	@echo "  make test               Run all tests"
	@echo "  make run-tests          Run all tests (alias)"
	@echo ""
	@echo "API Testing (requires running server):"
	@echo "  make list-services      List all available gRPC services"
	@echo "  make create-wheel       Create a test wheel"
	@echo "  make get-wheel          Get a specific wheel (requires WHEEL_ID)"
	@echo "  make list-wheels        List all wheels"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean              Clean build artifacts"
	@echo ""

build:
	bazel build //...

build-server:
	bazel build //backend/cmd/server:server

run-server:
	bazel run //backend/cmd/server:server

test:
	bazel test //...

run-tests: test

list-services:
	@command -v grpcurl >/dev/null 2>&1 || { echo "grpcurl not found. Installing..."; go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest; }
	grpcurl -plaintext localhost:50051 list

create-wheel:
	@command -v grpcurl >/dev/null 2>&1 || { echo "grpcurl not found. Installing..."; go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest; }
	grpcurl -plaintext -H "x-user-id: test-user" \
		-d '{"name": "Test Wheel", "initial_items": ["Red", "Blue", "Green", "Yellow"]}' \
		localhost:50051 spinwheel.v1.WheelService/CreateWheel

get-wheel:
	@if [ -z "$(WHEEL_ID)" ]; then \
		echo "Error: WHEEL_ID is required"; \
		echo "Usage: make get-wheel WHEEL_ID=<wheel-id>"; \
		exit 1; \
	fi
	@command -v grpcurl >/dev/null 2>&1 || { echo "grpcurl not found. Installing..."; go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest; }
	grpcurl -plaintext -H "x-user-id: test-user" \
		-d "{\"id\": \"$(WHEEL_ID)\"}" \
		localhost:50051 spinwheel.v1.WheelService/GetWheel

list-wheels:
	@command -v grpcurl >/dev/null 2>&1 || { echo "grpcurl not found. Installing..."; go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest; }
	grpcurl -plaintext -H "x-user-id: test-user" \
		-d '{"page_size": 10}' \
		localhost:50051 spinwheel.v1.WheelService/ListWheels

clean:
	bazel clean
	rm -rf bazel-*
