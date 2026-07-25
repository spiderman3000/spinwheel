# --- Stage 1: Build ---
FROM golang:1.23-bookworm AS builder

# Install protoc and git
RUN apt-get update && apt-get install -y \
    protobuf-compiler \
    git \
    && rm -rf /var/lib/apt/lists/*

# Install Go proto plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

WORKDIR /src
COPY . .

# Generate Go code from proto files
RUN mkdir -p backend/gen/proto/spinwheel/v1
RUN protoc \
    --go_out=backend --go_opt=module=spinwheel/backend \
    --go-grpc_out=backend --go-grpc_opt=module=spinwheel/backend \
    -I . \
    idl/proto/spinwheel/v1/wheel.proto \
    idl/proto/spinwheel/v1/service.proto

# Fix module path: imports use "spinwheel/backend/..." but go.mod says "module spinwheel"
RUN sed -i 's/^module spinwheel$/module spinwheel\/backend/' backend/go.mod && \
    cat backend/go.mod

# Build
RUN cd backend && go mod tidy && CGO_ENABLED=0 go build -o /server ./cmd/server

# --- Stage 2: Runtime ---
FROM gcr.io/distroless/base-debian12

COPY --from=builder /server /server

ENV PORT=8080
EXPOSE 8080

CMD ["/server"]
