package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/engineervix/pseudoc/internal/config"
	"github.com/engineervix/pseudoc/internal/server"
)

// runServer handles the server command
func runServer(args []string) error {
	// Initialize server config with defaults
	cfg := config.DefaultServerConfig()

	// Parse server-specific options
	if err := parseServerOptions(args, &cfg); err != nil {
		return enhanceError(err, "parsing server options")
	}

	// Validate server configuration
	if err := cfg.ValidateServer(); err != nil {
		return enhanceError(err, "validating server configuration")
	}

	// Create and start server
	srv := server.New(cfg)

	fmt.Printf("pseudoc server starting...\n")
	fmt.Printf("Address: %s\n", cfg.GetAddress())
	fmt.Printf("API Base: %s/api/v1\n", cfg.GetBaseURL())
	fmt.Printf("Health: %s/health\n", cfg.GetBaseURL())

	if cfg.EnableMetrics {
		fmt.Printf("Metrics: %s/metrics\n", cfg.GetBaseURL())
	}

	fmt.Printf("Configuration:\n")
	fmt.Printf("   • CORS: ")
	if len(cfg.CORSAllowedOrigins) > 0 {
		fmt.Printf("enabled for %s\n", strings.Join(cfg.CORSAllowedOrigins, ", "))
	} else {
		fmt.Println("disabled")
	}
	fmt.Printf("   • Rate limit: ")
	if cfg.RateLimit > 0 {
		fmt.Printf("%d req/min\n", cfg.RateLimit)
	} else {
		fmt.Printf("disabled\n")
	}
	fmt.Printf("   • Request timeout: %v\n", cfg.RequestTimeout)
	fmt.Printf("   • Max file size: %d MB\n", cfg.MaxFileSize/(1024*1024))
	fmt.Printf("   • Log level: %s\n", cfg.LogLevel)
	fmt.Printf("   • Logging: %v\n", cfg.EnableLogging)

	fmt.Printf("\nExample API calls:\n")
	fmt.Printf("   curl \"%s/api/v1/generate/pdf\"\n", cfg.GetBaseURL())
	fmt.Printf("   curl \"%s/api/v1/generate/docx?pages=3&seed=42\"\n", cfg.GetBaseURL())
	fmt.Printf("   curl -X POST \"%s/api/v1/generate\" \\\n", cfg.GetBaseURL())
	fmt.Printf("        -H \"Content-Type: application/json\" \\\n")
	fmt.Printf("        -d '{\"type\":\"pdf\",\"pages\":2,\"filename\":\"test\"}'\n")
	fmt.Printf("\nPress Ctrl+C to stop the server\n")
	fmt.Println(strings.Repeat("─", 50))

	// Start will block until shutdown signal is received
	return srv.Start()
}

// parseServerOptions parses command line options for the server
func parseServerOptions(args []string, cfg *config.ServerConfig) error {
	// Create a new FlagSet for server options
	flagSet := flag.NewFlagSet("server", flag.ContinueOnError)
	flagSet.Usage = func() {}

	// Server connection settings
	flagSet.StringVar(&cfg.Host, "host", cfg.Host, "Server host address")
	flagSet.IntVar(&cfg.Port, "port", cfg.Port, "Server port")
	flagSet.StringVar(&cfg.Environment, "env", cfg.Environment, "Server environment (development or production)")

	// Feature toggles
	var corsOrigins string
	flagSet.StringVar(&corsOrigins, "cors-allowed-origins", strings.Join(cfg.CORSAllowedOrigins, ","), "Comma-separated list of allowed CORS origins")
	flagSet.BoolVar(&cfg.EnableLogging, "logging", cfg.EnableLogging, "Enable request logging")
	flagSet.BoolVar(&cfg.EnableMetrics, "metrics", cfg.EnableMetrics, "Enable metrics endpoint")

	// Limits and timeouts
	flagSet.IntVar(&cfg.RateLimit, "rate-limit", cfg.RateLimit, "Requests per minute (0 = no limit)")
	flagSet.Int64Var(&cfg.MaxFileSize, "max-file-size", cfg.MaxFileSize, "Maximum file size in bytes")

	// Parse timeout as string then convert
	var timeoutStr string
	flagSet.StringVar(&timeoutStr, "timeout", cfg.RequestTimeout.String(), "Request timeout (e.g., 30s, 5m)")
	flagSet.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Log level (debug, info, warn, error)")

	// Parse the flags
	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("invalid server flag: %w\n\nUse 'pseudoc help' to see available options", err)
	}

	// Handle CORS origins
	if corsOrigins != "" {
		cfg.CORSAllowedOrigins = strings.Split(corsOrigins, ",")
	} else {
		cfg.CORSAllowedOrigins = []string{}
	}

	// Parse timeout if provided
	if timeoutStr != cfg.RequestTimeout.String() {
		if timeout, err := time.ParseDuration(timeoutStr); err != nil {
			return fmt.Errorf("invalid timeout format '%s': %w\n\nExamples: 30s, 5m, 1h", timeoutStr, err)
		} else {
			cfg.RequestTimeout = timeout
		}
	}

	return nil
}

// printServerUsage prints usage information for the server command
func printServerUsage() {
	fmt.Println("pseudoc serve - Start HTTP API server")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  pseudoc serve [options]")
	fmt.Println()
	fmt.Println("SERVER OPTIONS:")
	fmt.Println("  --host HOST                  Server host address (default: localhost)")
	fmt.Println("  --port PORT                  Server port (default: 8080)")
	fmt.Println("  --env ENV                    Server environment: development, production (default: development)")
	fmt.Println("  --timeout DURATION           Request timeout (default: 30s)")
	fmt.Println("  --max-file-size BYTES        Maximum generated file size (default: 104857600 = 100MB)")
	fmt.Println("  --rate-limit N               Requests per minute (default: 60, 0 = no limit)")
	fmt.Println("  --log-level LEVEL            Log level: debug, info, warn, error (default: info)")
	fmt.Println()
	fmt.Println("FEATURE FLAGS:")
	fmt.Println("  --cors-allowed-origins LIST  Comma-separated list of allowed CORS origins (default: \"*\")")
	fmt.Println("  --logging=true/false         Enable request logging (default: true)")
	fmt.Println("  --metrics=true/false         Enable metrics endpoint (default: false)")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Start server with defaults")
	fmt.Println("  pseudoc serve")
	fmt.Println()
	fmt.Println("  # Start on all interfaces with custom port")
	fmt.Println("  pseudoc serve --host 0.0.0.0 --port 3000")
	fmt.Println()
	fmt.Println("  # Start in production mode")
	fmt.Println("  pseudoc serve --env production")
	fmt.Println()
	fmt.Println("  # Start with custom timeout and no rate limiting")
	fmt.Println("  pseudoc serve --timeout 60s --rate-limit 0")
	fmt.Println()
	fmt.Println("  # Start with metrics enabled and debug logging")
	fmt.Println("  pseudoc serve --metrics --log-level debug")
	fmt.Println()
	fmt.Println("  # Start with higher rate limit for development")
	fmt.Println("  pseudoc serve --rate-limit 300 --max-file-size 200000000") // 200MB
	fmt.Println()
	fmt.Println("ENDPOINTS:")
	fmt.Println("  GET  /health                         Health check")
	fmt.Println("  GET  /api/v1/info                    Server information")
	fmt.Println("  GET  /api/v1/generate/{type}         Generate document with query params")
	fmt.Println("       ?pages=N&sheets=N&seed=N&filename=NAME")
	fmt.Println("  POST /api/v1/generate                Generate document with JSON body")
	fmt.Println("       {\"type\":\"pdf\",\"pages\":3,\"seed\":42}")
	fmt.Println("  GET  /metrics                        Server metrics (if --metrics enabled)")
	fmt.Println()
	fmt.Println("SUPPORTED DOCUMENT TYPES:")
	fmt.Println("  • pdf        - PDF documents")
	fmt.Println("  • docx       - Microsoft Word documents")
	fmt.Println("  • xlsx       - Microsoft Excel spreadsheets")
	fmt.Println("  • random     - Randomly selected document type")
	fmt.Println()
	fmt.Println("API EXAMPLES:")
	fmt.Println("  # Generate a simple PDF")
	fmt.Println("  curl \"http://localhost:8080/api/v1/generate/pdf\" -o document.pdf")
	fmt.Println()
	fmt.Println("  # Generate Word doc with 3 pages and custom filename")
	fmt.Println("  curl \"http://localhost:8080/api/v1/generate/docx?pages=3&filename=report\" -o report.docx")
	fmt.Println()
	fmt.Println("  # Generate Excel with JSON payload")
	fmt.Println("  curl -X POST \"http://localhost:8080/api/v1/generate\" \\")
	fmt.Println("       -H \"Content-Type: application/json\" \\")
	fmt.Println("       -d '{\"type\":\"xlsx\",\"sheets\":5,\"seed\":42}' -o data.xlsx")
	fmt.Println()
	fmt.Println("  # Get server info")
	fmt.Println("  curl \"http://localhost:8080/api/v1/info\" | jq")
	fmt.Println()
	fmt.Printf("For more information, visit: https://github.com/engineervix/pseudoc\n")
}
