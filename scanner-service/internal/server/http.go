package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/internal/config"
	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HTTPServer represents an HTTP server
type HTTPServer struct {
	server *http.Server
	router *gin.Engine
	logger *logger.Logger
	config config.HTTPServerConfig
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(cfg config.HTTPServerConfig, log *logger.Logger) *HTTPServer {
	// Set Gin mode
	if cfg.Port == 0 {
		cfg.Port = 8081
	}

	// Create router
	router := gin.New()

	// Create server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return &HTTPServer{
		server: server,
		router: router,
		logger: log,
		config: cfg,
	}
}

// Router returns the Gin router
func (s *HTTPServer) Router() *gin.Engine {
	return s.router
}

// Start starts the HTTP server
func (s *HTTPServer) Start() error {
	s.logger.Info("Starting HTTP server", zap.Int("port", s.config.Port))
	return s.server.ListenAndServe()
}

// Stop stops the HTTP server
func (s *HTTPServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server")
	return s.server.Shutdown(ctx)
}

// RegisterRoutes registers all HTTP routes
func (s *HTTPServer) RegisterRoutes(registerFunc func(router *gin.Engine)) {
	registerFunc(s.router)
}

// SetupMiddleware sets up common middleware
func (s *HTTPServer) SetupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Logger middleware
	s.router.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if query != "" {
			path = path + "?" + query
		}

		s.logger.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
		)
	})

	// CORS middleware
	s.router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
}
