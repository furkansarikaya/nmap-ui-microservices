package config

import "time"

// Config represents the application configuration
type Config struct {
	App     AppConfig
	Server  ServerConfig
	Nmap    NmapConfig
	Log     LogConfig
	Storage StorageConfig
}

// AppConfig contains application metadata
type AppConfig struct {
	Name    string
	Version string
}

// ServerConfig contains server configuration
type ServerConfig struct {
	HTTP HTTPServerConfig
	GRPC GRPCServerConfig
}

// HTTPServerConfig contains HTTP server configuration
type HTTPServerConfig struct {
	Port         int
	Timeout      time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// GRPCServerConfig contains gRPC server configuration
type GRPCServerConfig struct {
	Port    int
	Timeout time.Duration
}

// NmapConfig contains nmap configuration
type NmapConfig struct {
	Path               string
	Timeout            time.Duration
	MaxConcurrentScans int
}

// LogConfig contains logging configuration
type LogConfig struct {
	Level  string
	Format string
	Output string
}

// StorageConfig contains storage configuration
type StorageConfig struct {
	Type            string
	RetentionPeriod time.Duration
}
