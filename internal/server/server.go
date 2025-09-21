package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"github.com/engineervix/pseudoc/internal/config"
	"github.com/engineervix/pseudoc/internal/server/handlers"
)

// Server wraps the Echo server with configuration and metrics
type Server struct {
	echo    *echo.Echo
	config  config.ServerConfig
	metrics *ServerMetrics
}

// ServerMetrics tracks server performance metrics with atomic operations
type ServerMetrics struct {
	RequestCount       int64
	DocumentsGenerated map[string]*int64 // Changed to pointers for atomic operations
	ErrorCount         int64
	StartTime          time.Time
}

// NewServerMetrics creates a new ServerMetrics instance
func NewServerMetrics() *ServerMetrics {
	return &ServerMetrics{
		DocumentsGenerated: map[string]*int64{
			"pdf":  new(int64),
			"docx": new(int64),
			"xlsx": new(int64),
		},
		StartTime: time.Now(),
	}
}

// IncrementDocumentCount increments the document generation counter for a specific type
func (sm *ServerMetrics) IncrementDocumentCount(docType string) {
	if counter, exists := sm.DocumentsGenerated[docType]; exists {
		atomic.AddInt64(counter, 1)
	}
}

// GetDocumentCounts returns a copy of document counts for safe reading
func (sm *ServerMetrics) GetDocumentCounts() map[string]int64 {
	result := make(map[string]int64)
	for docType, counter := range sm.DocumentsGenerated {
		result[docType] = atomic.LoadInt64(counter)
	}
	return result
}

// GetRequestCount returns the current request count
func (sm *ServerMetrics) GetRequestCount() int64 {
	return atomic.LoadInt64(&sm.RequestCount)
}

// GetErrorCount returns the current error count
func (sm *ServerMetrics) GetErrorCount() int64 {
	return atomic.LoadInt64(&sm.ErrorCount)
}

// GetStartTime returns the server start time
func (sm *ServerMetrics) GetStartTime() time.Time {
	return sm.StartTime
}

// New creates a new server instance with the given configuration
func New(cfg config.ServerConfig) *Server {
	e := echo.New()

	// Configure Echo
	e.HideBanner = true
	e.HidePort = true

	// Custom error handler
	e.HTTPErrorHandler = customHTTPErrorHandler

	return &Server{
		echo:    e,
		config:  cfg,
		metrics: NewServerMetrics(),
	}
}

// setupMiddleware configures the middleware stack
func (s *Server) setupMiddleware() {
	// Request ID middleware - should be first for proper logging
	s.echo.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return generateRequestID()
		},
	}))

	// Request timeout middleware
	s.echo.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout:      s.config.RequestTimeout,
		ErrorMessage: "Request timeout - the server took too long to generate the document",
	}))

	// CORS middleware (if enabled)
	if s.config.EnableCORS {
		s.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodOptions},
			AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderXRequestID},
			ExposeHeaders:    []string{echo.HeaderXRequestID},
			AllowCredentials: false,
		}))
	}

	// Rate limiting middleware (if enabled)
	if s.config.RateLimit > 0 {
		s.echo.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
			rate.Limit(float64(s.config.RateLimit) / 60.0), // Convert per-minute to per-second
		)))
	}

	// Metrics middleware - tracks requests
	s.echo.Use(s.metricsMiddleware())

	// Request logging middleware (if enabled)
	if s.config.EnableLogging {
		logFormat := `{"time":"${time_rfc3339_nano}","method":"${method}","uri":"${uri}",` +
			`"status":${status},"latency_human":"${latency_human}","bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n"

		if s.config.LogLevel == "debug" {
			logFormat = `{"time":"${time_rfc3339_nano}","id":"${id}","method":"${method}","uri":"${uri}",` +
				`"status":${status},"latency_human":"${latency_human}","bytes_in":${bytes_in},"bytes_out":${bytes_out},"error":"${error}"}` + "\n"
		}

		s.echo.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: logFormat,
		}))
	}

	// Recovery middleware
	s.echo.Use(middleware.Recover())

	// Body limit middleware to prevent huge requests
	s.echo.Use(middleware.BodyLimit(fmt.Sprintf("%d", s.config.MaxFileSize)))

	// Security headers middleware
	s.echo.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000, // 1 year
		ContentSecurityPolicy: "default-src 'self'",
	}))
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.echo.GET("/health", handlers.NewHealthHandler())

	// API v1 group
	api := s.echo.Group("/api/v1")

	// Document generation endpoints
	docHandler := handlers.NewDocumentHandler(&s.config, s.metrics)

	// GET endpoint for simple document generation
	api.GET("/generate/:type", docHandler.GenerateGET)

	// POST endpoint for advanced document generation
	api.POST("/generate", docHandler.GeneratePOST)

	// Server info endpoint
	api.GET("/info", handlers.NewInfoHandler(&s.config))

	// Metrics endpoint (if enabled)
	if s.config.EnableMetrics {
		s.echo.GET("/metrics", handlers.NewMetricsHandler(s.metrics))
	}

	// CORS preflight handling for the generate endpoint
	if s.config.EnableCORS {
		api.OPTIONS("/generate", func(c echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		})
		api.OPTIONS("/generate/:type", func(c echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		})
	}
}

// metricsMiddleware tracks request metrics
func (s *Server) metricsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Increment request counter
			atomic.AddInt64(&s.metrics.RequestCount, 1)

			// Call the next handler
			err := next(c)

			// Track errors
			if err != nil {
				atomic.AddInt64(&s.metrics.ErrorCount, 1)
			} else if c.Response().Status >= 400 {
				atomic.AddInt64(&s.metrics.ErrorCount, 1)
			}

			return err
		}
	}
}

// customHTTPErrorHandler provides consistent error responses
func customHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var (
		code    = http.StatusInternalServerError
		message = "Internal server error"
	)

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = fmt.Sprintf("%v", he.Message)
	}

	// Log the error
	c.Logger().Errorf("HTTP error: %v", err)

	// Don't leak internal errors to clients
	if code == http.StatusInternalServerError {
		message = "Internal server error"
	}

	errorResponse := handlers.ErrorResponse{
		Error:   http.StatusText(code),
		Message: message,
		Code:    code,
	}

	// Send JSON error response
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead {
			c.NoContent(code)
		} else {
			c.JSON(code, errorResponse)
		}
	}
}

// generateRequestID creates a simple request ID
func generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix()%1000)
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
