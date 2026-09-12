package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"

	spinhttp "spinwheel/backend/internal/http"
	"spinwheel/backend/internal/repository"
	"spinwheel/backend/internal/service"
	"spinwheel/backend/pkg/logger"
	"spinwheel/backend/pkg/models"
)

// NOTE (HTTP-only decision, 2026-09-12): the gRPC service is NOT served —
// exposing it is backlog. internal/handler stays built + unit-tested;
// mount it here (port, reflection, interceptors) when that lands.

func main() {
	env := getEnv("ENVIRONMENT", "development")
	if err := logger.Init(env); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	hmacSecret := strings.TrimSpace(os.Getenv("HMAC_SECRET"))
	if hmacSecret == "" {
		logger.Log.Fatal("HMAC_SECRET is required (spins fail closed without it)")
	}

	// Repository: Neon Postgres when DATABASE_URL (pooled) is set,
	// otherwise an in-memory store seeded with a demo wheel for local dev.
	var (
		repo repository.WheelRepository
		pg   *repository.PostgresWheelRepository
	)
	if dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL")); dbURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var err error
		pg, err = repository.NewPostgresWheelRepository(ctx, dbURL)
		if err != nil {
			logger.Log.Fatal("failed to connect to Postgres", zap.Error(err))
		}
		defer pg.Close()
		repo = pg
		logger.Log.Info("using Postgres repository")
	} else {
		repo = repository.NewInMemoryWheelRepository()
		logger.Log.Warn("DATABASE_URL unset — using in-memory repository (data will not survive restarts)")
	}

	serv := service.NewWheelService(repo, hmacSecret)

	if pg == nil {
		seedDemoWheel(serv)
	}

	var corsOrigins []string
	if raw := strings.TrimSpace(os.Getenv("CORS_ORIGINS")); raw != "" {
		corsOrigins = strings.Split(raw, ",")
	}

	gateway := spinhttp.NewServer(spinhttp.Config{
		Service:        serv,
		Logger:         logger.Log,
		CORSOrigins:    corsOrigins,
		SecureCookies:  env == "production",
		RequestsPerMin: getEnvInt("REQUESTS_PER_MIN", 100),
	})

	port := getEnv("PORT", "8080") // Cloud Run injects PORT
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           gateway.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Log.Info("HTTP server listening", zap.String("port", port), zap.String("env", env))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Log.Info("shutting down")
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Fatal("graceful shutdown failed", zap.Error(err))
	}
	logger.Log.Info("server stopped")
}

// seedDemoWheel creates a starter wheel for local dev so the HTTP API is
// usable without a database or wheel-creation route. In-memory only.
func seedDemoWheel(serv service.WheelService) {
	wheel, err := serv.CreateWheel(&models.Wheel{
		Name: "Demo wheel",
		Items: []models.WheelItem{
			{Option: "Red", Color: "#ef4444", Weight: 1.0},
			{Option: "Green", Color: "#22c55e", Weight: 1.0},
			{Option: "Blue", Color: "#3b82f6", Weight: 1.0},
		},
	})
	if err != nil {
		logger.Log.Fatal("failed to seed demo wheel", zap.Error(err))
	}
	logger.Log.Info("seeded demo wheel",
		zap.String("wheel_id", wheel.ID),
		zap.String("items_hash", service.ComputeItemsHash(wheel.Items)),
	)
}

func getEnv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if raw := strings.TrimSpace(os.Getenv(key)); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			return v
		}
		logger.Log.Warn("invalid int env, using default",
			zap.String("key", key), zap.String("value", raw), zap.Int("default", def))
	}
	return def
}
