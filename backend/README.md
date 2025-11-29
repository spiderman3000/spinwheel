# Spinwheel Backend

This document provides instructions on how to build, start, and test the backend service for the Spinwheel application.

## Building the Backend

*   **Prerequisites:** Make sure you have [Bazel](https://bazel.build/install) installed on your system.
*   **Build Command:** To build the backend server, run the following command from the root of the project:

    ```bash
    bazel build //backend/cmd/server:server
    ```

## Starting the Backend Locally

*   **Run Command:** You can start the backend server directly using Bazel with the following command:

    ```bash
    bazel run //backend/cmd/server:server
    ```

*   **Endpoints:**
    *   The gRPC server will be available on port `:50051`.
    *   The REST gateway will be available on port `:8080`.

## Testing the Backend

*   **Test Command:** To run the tests for the backend, use the following command:

    ```bash
    bazel test //...
    ```

## Available APIs

The backend exposes a RESTful API via a gRPC-gateway.

*   **Create a Wheel:**
    *   `POST /v1/wheels`
    *   **Body:**
        ```json
        {
          "name": "My Wheel",
          "items": [
            { "text": "Item 1", "color": "#ff0000" },
            { "text": "Item 2", "color": "#00ff00" }
          ]
        }
        ```

*   **Get a Wheel:**
    *   `GET /v1/wheels/{id}`

*   **Update a Wheel:**
    *   `PUT /v1/wheels/{id}`
    *   **Body:**
        ```json
        {
          "name": "Updated Wheel Name",
          "items": [
            { "id": "item-id-1", "text": "Updated Item 1", "color": "#ff0000" },
            { "text": "New Item 3", "color": "#0000ff" }
          ]
        }
        ```

*   **Delete a Wheel:**
    *   `DELETE /v1/wheels/{id}`

*   **Spin a Wheel:**
    *   `POST /v1/wheels/{id}:spin`
    *   **Returns:** The winning item.

## Code Structure

The backend code is organized using a clean architecture approach:

*   `cmd/server/main.go`: The entry point for the backend server.
*   `internal/handler`: Contains the gRPC handlers that process incoming requests.
*   `internal/service`: Contains the business logic for the application.
*   `internal/repository`: The data access layer. Currently uses an in-memory store.
*   `pkg/models`: Defines the core data structures (structs) used in the application.
*   `gen/proto`: Contains the generated Go code from the `.proto` files.

## Important Libraries and Package Managers

*   **[Bazel](https://bazel.build/):** The build system used for the project. It handles dependency management and provides a consistent build environment.
*   **[gRPC](https://grpc.io/):** A high-performance, open-source universal RPC framework. Used for communication between the frontend and backend.
*   **[gRPC-Gateway](https://github.com/grpc-ecosystem/grpc-gateway):** A plugin for `protoc` that generates a reverse-proxy server which translates a RESTful JSON API into gRPC.
*   **[Protocol Buffers (Protobuf)](https://developers.google.com/protocol-buffers):** The Interface Definition Language (IDL) used to define the API contract between the frontend and backend.

# spinwheel
Spinwheel provides a gamified experience like wheel for fortune

## Backend

The backend is a Go-based application built with Bazel. It uses a 3-layer architecture (Handler -> Service -> Repository) and supports both gRPC and REST protocols.

### Development Environment Setup

The backend is built and managed using Bazel. If you don't have Bazel installed, you can install it using the following commands:

**Install Bazel:**

```bash
sudo apt-get update && sudo apt-get install -y curl gnupg
curl https://bazel.build/bazel-release.pub.gpg | sudo apt-key add -
echo "deb [arch=amd64] https://storage.googleapis.com/bazel-apt stable jdk1.8" | sudo tee /etc/apt/sources.list.d/bazel.list
sudo apt-get update && sudo apt-get install -y bazel
```

The project requires a specific version of bazel, specified in the `.bazelversion` file. After installing bazel, you may need to install the correct version. For example, if the `.bazelversion` file specifies `7.1.1`, you can install it with:

```bash
sudo apt-get install -y bazel-7.1.1
```

### Building, Testing, and Running

**Build the backend:**

```bash
bazel build //...
```

**Build the backend server only:**

```bash
bazel build //backend/cmd/server:server
```

**Run backend tests:**

```bash
bazel test //...
```

**Run the backend server:**

```bash
bazel run //backend/cmd/server:server
```

The gRPC server will start on `localhost:50051`.

### Testing the Backend API

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

**Get a wheel (replace with actual ID):**

```bash
grpcurl -plaintext -H "x-user-id: test-user" \
  -d '{"id": "wheel-id"}' \
  localhost:50051 spinwheel.v1.WheelService/GetWheel
```

**List all wheels:**

```bash
grpcurl -plaintext -H "x-user-id: test-user" \
  -d '{"page_size": 10}' \
  localhost:50051 spinwheel.v1.WheelService/ListWheels
```

**Note:** The `x-user-id` header is required by the UserIDInterceptor middleware.

## Frontend

The frontend is a React application built with Vite, TypeScript, and Tailwind CSS. It uses the `react-custom-roulette` library for the wheel component.

- [react-custom-roulette](https://www.npmjs.com/package/react-custom-roulette?activeTab=readme)


## Security Considerations

### For Development
- Never commit `.env` files or secrets
- Use `.env.local` for local development
- The `X-User-Id` header is for development only

### For Production
- Replace `X-User-Id` with proper JWT authentication
- Enable HTTPS/TLS for all connections
- Use environment variables for all sensitive configuration
- Enable rate limiting (100 req/min per user)
- Rotate any exposed credentials immediately
- Review CORS settings for production domains

### Reporting Security Issues
Please report security vulnerabilities to [your-email] or use GitHub Security Advisories.