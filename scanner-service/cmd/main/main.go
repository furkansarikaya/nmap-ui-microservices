package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/internal/config"
	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/internal/features/scan/adapters"
	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/internal/features/scan/domain"
	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/internal/features/scan/handlers"
	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/internal/features/scan/repository"
	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/internal/server"
	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.NewLogger(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		Output: cfg.Log.Output,
	})
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting Scanner Service",
		zap.String("name", cfg.App.Name),
		zap.String("version", cfg.App.Version),
	)

	// Initialize nmap adapter
	nmapAdapter := adapters.NewNmapAdapter(cfg.Nmap.Path, log)

	// Check if nmap is available
	if !nmapAdapter.IsAvailable() {
		log.Fatal("Nmap is not available. Please install nmap and try again.")
	}

	// Initialize repository
	scanRepo := repository.NewMemoryScanRepository(log, cfg.Storage.RetentionPeriod)

	// Initialize scan service
	scanService := domain.NewScanService(nmapAdapter, scanRepo, log, cfg.Nmap.MaxConcurrentScans)

	// Initialize HTTP server
	httpServer := server.NewHTTPServer(cfg.Server.HTTP, log)
	httpServer.SetupMiddleware()

	// Initialize scan handler
	scanHandler := handlers.NewScanHandler(scanService, log)

	// Register routes
	httpServer.RegisterRoutes(func(router *gin.Engine) {
		// Register scan handler routes
		scanHandler.RegisterRoutes(router)
	})

	// Initialize gRPC server
	grpcServer, err := server.NewGRPCServer(cfg.Server.GRPC, log)
	if err != nil {
		log.Fatal("Failed to create gRPC server", zap.Error(err))
	}

	// Start servers in separate goroutines
	go func() {
		if err := httpServer.Start(); err != nil {
			log.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	go func() {
		if err := grpcServer.Start(); err != nil {
			log.Fatal("Failed to start gRPC server", zap.Error(err))
		}
	}()

	log.Info("Servers started",
		zap.Int("http_port", cfg.Server.HTTP.Port),
		zap.Int("grpc_port", cfg.Server.GRPC.Port),
	)

	// Wait for interrupt signal to gracefully shut down the servers
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down servers...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Stop gRPC server
	grpcServer.Stop()

	// Stop HTTP server
	if err := httpServer.Stop(ctx); err != nil {
		log.Error("Failed to gracefully shutdown HTTP server", zap.Error(err))
	}

	log.Info("Servers successfully shutdown")
}
