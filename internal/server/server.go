package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"github.com/engineervix/pseudoc/internal/config"
	"github.com/engineervix/pseudoc/internal/server/handlers"
)

// Server wraps the Echo server with configuration
type Server struct {
	echo   *echo.Echo
	config config.ServerConfig
}

// New creates a new server instance with the given configuration
func New(cfg config.ServerConfig) *Server {
	e := echo.New()

	// Configure Echo
	e.HideBanner = true
	e.HidePort = true

	return &Server{
		echo:   e,
		config: cfg,
	}
}

// setupMiddleware configures the middleware stack
func (s *Server) setupMiddleware() {
	// Request timeout middleware
	s.echo.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: s.config.RequestTimeout,
	}))

	// CORS middleware (if enabled)
	if s.config.EnableCORS {
		s.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		}))
	}

	// Rate limiting middleware (if enabled)
	if s.config.RateLimit > 0 {
		s.echo.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
			rate.Limit(float64(s.config.RateLimit) / 60.0), // Convert per-minute to per-second
		)))
	}

	// Request logging middleware (if enabled)
	if s.config.EnableLogging {
		s.echo.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: `{"time":"${time_rfc3339_nano}","method":"${method}","uri":"${uri}",` +
				`"status":${status},"latency_human":"${latency_human}","bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
		}))
	}

	// Recovery middleware
	s.echo.Use(middleware.Recover())

	// Body limit middleware to prevent huge requests
	s.echo.Use(middleware.BodyLimit(fmt.Sprintf("%d", s.config.MaxFileSize)))
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {
	// Health check endpoint (no /api prefix for simplicity)
	s.echo.GET("/health", handlers.NewHealthHandler())

	// API v1 group
	api := s.echo.Group("/api/v1")

	// Document generation endpoints
	docHandler := handlers.NewDocumentHandler(&s.config)
	api.GET("/generate/:type", docHandler.GenerateGET)
	api.POST("/generate", docHandler.GeneratePOST)

	// Server info endpoint
	api.GET("/info", handlers.NewInfoHandler(&s.config))

	// Metrics endpoint (if enabled)
	if s.config.EnableMetrics {
		s.echo.GET("/metrics", handlers.NewMetricsHandler())
	}
}

// Start starts the HTTP server with graceful shutdown
func (s *Server) Start() error {
	// Setup middleware and routes
	s.setupMiddleware()
	s.setupRoutes()

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in a goroutine
	go func() {
		address := s.config.GetAddress()
		if !s.config.EnableLogging {
			fmt.Printf("Starting server on %s\n", address)
			fmt.Printf("API available at: %s/api/v1\n", s.config.GetBaseURL())
			fmt.Printf("Health check: %s/health\n", s.config.GetBaseURL())
		}

		if err := s.echo.Start(address); err != nil && err != http.ErrServerClosed {
			s.echo.Logger.Fatal("Server startup failed: ", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-quit:
		if !s.config.EnableLogging {
			fmt.Printf("\nReceived signal: %v\n", sig)
			fmt.Println("Shutting down server...")
		}
	case <-ctx.Done():
		// Context cancelled externally
	}

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := s.echo.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	if !s.config.EnableLogging {
		fmt.Println("Server stopped gracefully")
	}

	return nil
}

// Stop stops the server gracefully
func (s *Server) Stop(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
