package config

import (
	"fmt"
	"time"
)

// ServerConfig wraps the core Config with server-specific settings
type ServerConfig struct {
	Config // Embedded core config (used as defaults for API requests)

	// Server settings
	Host        string // Server host (e.g., "0.0.0.0", "localhost")
	Port        int    // Server port (e.g., 8080)
	Environment string // Server environment (e.g., "development", "production")

	// API behavior
	CORSAllowedOrigins []string      // List of allowed CORS origins. If empty, CORS is disabled.
	RequestTimeout     time.Duration // Request timeout
	MaxFileSize        int64         // Maximum generated file size in bytes
	RateLimit          int           // Requests per minute (0 = no limit)
	EnableMetrics      bool          // Enable metrics endpoint

	// Logging
	EnableLogging bool   // Enable request logging
	LogLevel      string // Log level (debug, info, warn, error)

	// Security
	XFrameOptions         string // X-Frame-Options header
	ContentSecurityPolicy string // Content-Security-Policy header
}

// DefaultServerConfig returns a ServerConfig with sensible defaults
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Config:             DefaultConfig(),
		Host:               "localhost",
		Port:               8080,
		Environment:        "development",
		CORSAllowedOrigins: []string{}, // Default to no CORS - must be explicitly configured
		RequestTimeout:     30 * time.Second,
		MaxFileSize:        100 * 1024 * 1024, // 100MB
		RateLimit:          60,                // 60 requests per minute
		EnableMetrics:      false,
		EnableLogging:      true,
		LogLevel:           "info",
		// Security
		XFrameOptions:         "SAMEORIGIN",
		ContentSecurityPolicy: "default-src 'self'",
	}
}

// ValidateServer validates the server configuration
func (c *ServerConfig) ValidateServer() error {
	// Validate core config
	if err := c.Config.Validate(); err != nil {
		return err
	}

	// Validate server-specific settings
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
	}

	if c.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	if c.Environment != "development" && c.Environment != "production" {
		return fmt.Errorf("environment must be 'development' or 'production', got '%s'", c.Environment)
	}

	if c.RequestTimeout <= 0 {
		return fmt.Errorf("request timeout must be positive")
	}

	if c.MaxFileSize <= 0 {
		return fmt.Errorf("max file size must be positive")
	}

	if c.RateLimit < 0 {
		return fmt.Errorf("rate limit cannot be negative")
	}

	// Validate log level
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
		// Valid log levels
	default:
		return fmt.Errorf("invalid log level: %s", c.LogLevel)
	}

	return nil
}

// GetAddress returns the full address string for the server
func (c *ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetBaseURL returns the base URL for the server
func (c *ServerConfig) GetBaseURL() string {
	return fmt.Sprintf("http://%s", c.GetAddress())
}
