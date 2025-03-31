package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// LoadConfig loads configuration from file and environment variables
func LoadConfig() (*Config, error) {
	// Set default configuration file path
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("/etc/scanner-service")
	viper.AddConfigPath("$HOME/.scanner-service")

	// Read environment variables with prefix SCANNER_
	viper.SetEnvPrefix("SCANNER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, continue with defaults and env vars
			fmt.Println("Config file not found, using defaults and environment variables")
		} else {
			// Config file was found but another error occurred
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	config := &Config{}

	// App configuration
	config.App.Name = viper.GetString("app.name")
	config.App.Version = viper.GetString("app.version")

	// HTTP Server configuration
	config.Server.HTTP.Port = viper.GetInt("server.http.port")
	config.Server.HTTP.Timeout = viper.GetDuration("server.http.timeout")
	config.Server.HTTP.ReadTimeout = viper.GetDuration("server.http.read_timeout")
	config.Server.HTTP.WriteTimeout = viper.GetDuration("server.http.write_timeout")

	// gRPC Server configuration
	config.Server.GRPC.Port = viper.GetInt("server.grpc.port")
	config.Server.GRPC.Timeout = viper.GetDuration("server.grpc.timeout")

	// Nmap configuration
	config.Nmap.Path = viper.GetString("nmap.path")
	config.Nmap.Timeout = viper.GetDuration("nmap.timeout")
	config.Nmap.MaxConcurrentScans = viper.GetInt("nmap.max_concurrent_scans")

	// Logging configuration
	config.Log.Level = viper.GetString("log.level")
	config.Log.Format = viper.GetString("log.format")
	config.Log.Output = viper.GetString("log.output")

	// Storage configuration
	config.Storage.Type = viper.GetString("storage.type")
	config.Storage.RetentionPeriod = viper.GetDuration("storage.retention_period")

	// Set defaults if not provided
	setDefaults(config)

	return config, nil
}

// setDefaults sets default values for configuration if not provided
func setDefaults(config *Config) {
	// App defaults
	if config.App.Name == "" {
		config.App.Name = "scanner-service"
	}
	if config.App.Version == "" {
		config.App.Version = "0.1.0"
	}

	// HTTP Server defaults
	if config.Server.HTTP.Port == 0 {
		config.Server.HTTP.Port = 8081
	}
	if config.Server.HTTP.Timeout == 0 {
		config.Server.HTTP.Timeout = 30 * time.Second
	}
	if config.Server.HTTP.ReadTimeout == 0 {
		config.Server.HTTP.ReadTimeout = 15 * time.Second
	}
	if config.Server.HTTP.WriteTimeout == 0 {
		config.Server.HTTP.WriteTimeout = 15 * time.Second
	}

	// gRPC Server defaults
	if config.Server.GRPC.Port == 0 {
		config.Server.GRPC.Port = 9081
	}
	if config.Server.GRPC.Timeout == 0 {
		config.Server.GRPC.Timeout = 30 * time.Second
	}

	// Nmap defaults
	if config.Nmap.Path == "" {
		config.Nmap.Path = "nmap"
	}
	if config.Nmap.Timeout == 0 {
		config.Nmap.Timeout = 300 * time.Second
	}
	if config.Nmap.MaxConcurrentScans == 0 {
		config.Nmap.MaxConcurrentScans = 5
	}

	// Logging defaults
	if config.Log.Level == "" {
		config.Log.Level = "info"
	}
	if config.Log.Format == "" {
		config.Log.Format = "json"
	}
	if config.Log.Output == "" {
		config.Log.Output = "stdout"
	}

	// Storage defaults
	if config.Storage.Type == "" {
		config.Storage.Type = "memory"
	}
	if config.Storage.RetentionPeriod == 0 {
		config.Storage.RetentionPeriod = 168 * time.Hour // 7 days
	}
}
