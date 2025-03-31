package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/internal/config"
	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPCServer represents a gRPC server
type GRPCServer struct {
	server *grpc.Server
	config config.GRPCServerConfig
	logger *logger.Logger
	lis    net.Listener
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(cfg config.GRPCServerConfig, log *logger.Logger) (*GRPCServer, error) {
	// Create listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	// Create server with interceptors
	server := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor(log)),
	)

	// Enable reflection for grpcurl
	reflection.Register(server)

	return &GRPCServer{
		server: server,
		config: cfg,
		logger: log,
		lis:    lis,
	}, nil
}

// Start starts the gRPC server
func (s *GRPCServer) Start() error {
	s.logger.Info("Starting gRPC server", zap.Int("port", s.config.Port))
	return s.server.Serve(s.lis)
}

// Stop stops the gRPC server
func (s *GRPCServer) Stop() {
	s.logger.Info("Stopping gRPC server")
	s.server.GracefulStop()
}

// Server returns the underlying gRPC server
func (s *GRPCServer) Server() *grpc.Server {
	return s.server
}

// loggingInterceptor creates a logging interceptor for gRPC
func loggingInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		// Log the request
		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
			log.Error("gRPC request failed", fields...)
		} else {
			log.Info("gRPC request completed", fields...)
		}

		return resp, err
	}
}
