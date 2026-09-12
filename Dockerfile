# --- Stage 1: Build ---
FROM golang:1.23-bookworm AS builder

# Install git (needed for Go module downloads)
RUN apt-get update && apt-get install -y \
    git \
    && rm -rf /var/lib/apt/lists/*

# Pinned codegen toolchain (SPI-7 fragility fix): buf bundles protoc, so no
# apt `protobuf-compiler` version drift; plugin versions are pinned, never
# @latest. Matches buf.yaml / buf.gen.yaml at the repo root.
ARG BUF_VERSION=v1.32.2
ARG PROTOC_GEN_GO_VERSION=v1.34.2
ARG PROTOC_GEN_GO_GRPC_VERSION=v1.5.1
RUN go install github.com/bufbuild/buf/cmd/buf@${BUF_VERSION} \
    && go install google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOC_GEN_GO_VERSION} \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@${PROTOC_GEN_GO_GRPC_VERSION}
ENV PATH="/root/go/bin:${PATH}"

WORKDIR /src
COPY . .

# Generate Go code from proto files (reproducible: toolchain pinned above)
RUN buf generate

# Build (go.mod is `module spinwheel/backend`, imports resolve directly)
RUN cd backend && go mod tidy && CGO_ENABLED=0 go build -o /server ./cmd/server

# --- Stage 2: Runtime ---
FROM gcr.io/distroless/base-debian12

COPY --from=builder /server /server

ENV PORT=8080
EXPOSE 8080

CMD ["/server"]
