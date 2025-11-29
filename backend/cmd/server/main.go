package main

import (
	"log"
	"net"
	"os"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	spinwheelv1 "spinwheel/backend/gen/proto/spinwheel/v1"
	"spinwheel/backend/internal/handler"
	"spinwheel/backend/internal/middleware"
	"spinwheel/backend/internal/repository"
	"spinwheel/backend/internal/service"
	"spinwheel/backend/pkg/logger"
)

func main() {
	// Initialize logger
	if err := logger.Init("development"); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Log.Info("Starting Spinwheel server")

	// Create a new in-memory repository
	repo := repository.NewInMemoryWheelRepository()

	// Create a new service
	serv := service.NewWheelService(repo)

	// Create a new handler
	hand := handler.NewWheelHandler(serv)

	// Create gRPC server with interceptors
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(middleware.UserIDInterceptor()),
		grpc.Creds(insecure.NewCredentials()),
		// Add connection timeouts
		grpc.ConnectionTimeout(time.Second * 30),
		grpc.MaxRecvMsgSize(1024 * 1024), // 1MB max message size
		grpc.MaxSendMsgSize(1024 * 1024),
	}
	grpcServer := grpc.NewServer(opts...)

	// Register the handler with the gRPC server
	spinwheelv1.RegisterWheelServiceServer(grpcServer, hand)

	// Only register reflection in development
	if os.Getenv("ENVIRONMENT") != "production" {
		reflection.Register(grpcServer)
		logger.Log.Info("gRPC reflection enabled (development mode)")
	}

	// Get PORT from environment or default to 8080 (Cloud Run standard)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Log.Info("gRPC server starting", zap.String("port", port))

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Log.Fatal("failed to listen", zap.Error(err))
	}

	// Start the gRPC server
	logger.Log.Info("gRPC server listening", zap.String("port", port))
	if err := grpcServer.Serve(lis); err != nil {
		logger.Log.Fatal("failed to serve gRPC", zap.Error(err))
	}
}
