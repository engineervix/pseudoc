package server

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/engineervix/pseudoc/internal/config"
	"github.com/engineervix/pseudoc/internal/server/handlers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
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
	Performance        *PerformanceTracker
}

// PerformanceTracker tracks response times and request rates
type PerformanceTracker struct {
	mutex         sync.RWMutex
	responseTimes []time.Duration
	requestTimes  []time.Time
	writeIndex    int
	maxSamples    int
	totalRequests int64
}

// NewPerformanceTracker creates a new PerformanceTracker instance
func NewPerformanceTracker() *PerformanceTracker {
	return &PerformanceTracker{
		responseTimes: make([]time.Duration, config.MaxPerformanceSamples),
		requestTimes:  make([]time.Time, config.MaxPerformanceSamples),
		maxSamples:    config.MaxPerformanceSamples,
	}
}

// AddResponseTime records a response time and request timestamp
func (pt *PerformanceTracker) AddResponseTime(duration time.Duration) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	pt.responseTimes[pt.writeIndex] = duration
	pt.requestTimes[pt.writeIndex] = time.Now()
	pt.writeIndex = (pt.writeIndex + 1) % pt.maxSamples
	atomic.AddInt64(&pt.totalRequests, 1)
}

// GetAverageResponseTime calculates the average response time from stored samples
func (pt *PerformanceTracker) GetAverageResponseTime() time.Duration {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()

	totalRequests := atomic.LoadInt64(&pt.totalRequests)
	if totalRequests == 0 {
		return 0
	}

	var totalDuration time.Duration
	var count int64

	// Calculate how many valid samples we have
	samples := int(totalRequests)
	if samples > pt.maxSamples {
		samples = pt.maxSamples
	}

	// Sum up all valid response times
	for i := 0; i < samples; i++ {
		if pt.responseTimes[i] > 0 {
			totalDuration += pt.responseTimes[i]
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return totalDuration / time.Duration(count)
}

// GetRequestsPerMinute calculates requests per minute based on recent activity
func (pt *PerformanceTracker) GetRequestsPerMinute() float64 {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()

	now := time.Now()
	cutoff := now.Add(-config.RequestsPerMinuteWindow)

	var count int
	totalRequests := atomic.LoadInt64(&pt.totalRequests)

	// Count requests within the last minute
	samples := int(totalRequests)
	if samples > pt.maxSamples {
		samples = pt.maxSamples
	}

	for i := 0; i < samples; i++ {
		if !pt.requestTimes[i].IsZero() && pt.requestTimes[i].After(cutoff) {
			count++
		}
	}

	return float64(count)
}

// HasSufficientSamples returns true if we have enough samples for reliable metrics
func (pt *PerformanceTracker) HasSufficientSamples() bool {
	return atomic.LoadInt64(&pt.totalRequests) >= config.MinSamplesForMetrics
}

// NewServerMetrics creates a new ServerMetrics instance
func NewServerMetrics() *ServerMetrics {
	return &ServerMetrics{
		DocumentsGenerated: map[string]*int64{
			"pdf":  new(int64),
			"docx": new(int64),
			"xlsx": new(int64),
		},
		StartTime:   time.Now(),
		Performance: NewPerformanceTracker(),
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

// GetAverageResponseTime returns the average response time from the performance tracker
func (sm *ServerMetrics) GetAverageResponseTime() time.Duration {
	return sm.Performance.GetAverageResponseTime()
}

// GetRequestsPerMinute returns the current requests per minute rate
func (sm *ServerMetrics) GetRequestsPerMinute() float64 {
	return sm.Performance.GetRequestsPerMinute()
}

// HasSufficientSamples returns true if we have enough samples for reliable metrics
func (sm *ServerMetrics) HasSufficientSamples() bool {
	return sm.Performance.HasSufficientSamples()
}

// isLogLevelEnabled checks if a log with a given level should be shown based on server config.
func isLogLevelEnabled(level, configLevel string) bool {
	levels := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
	}
	configPriority, ok := levels[configLevel]
	if !ok {
		configPriority = levels["info"] // Default to info
	}
	levelPriority := levels[level]
	return levelPriority >= configPriority
}

// New creates a new server instance with the given configuration
func New(cfg config.ServerConfig) *Server {
	e := echo.New()

	// Configure Echo
	e.HideBanner = true
	e.HidePort = true

	// Custom IP extractor to handle reverse proxies (e.g., Cloudflare, Traefik)
	e.IPExtractor = func(req *http.Request) string {
		// Cloudflare header
		if ip := req.Header.Get("CF-Connecting-IP"); ip != "" {
			return ip
		}

		// Other common proxy headers
		if ip := req.Header.Get("X-Real-IP"); ip != "" {
			return ip
		}

		// Use Echo's robust X-Forwarded-For extractor, trusting common private networks.
		// This correctly handles proxy chains and avoids spoofed headers.
		extractor := echo.ExtractIPFromXFFHeader(
			echo.TrustLoopback(true),
			echo.TrustLinkLocal(true),
			echo.TrustPrivateNet(true),
		)
		if ip := extractor(req); ip != "" {
			return ip
		}

		// Default to RemoteAddr as the final fallback
		return req.RemoteAddr
	}

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
	if len(s.config.CORSAllowedOrigins) > 0 {
		s.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     s.config.CORSAllowedOrigins,
			AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodOptions},
			AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderXRequestID},
			ExposeHeaders:    []string{echo.HeaderXRequestID},
			AllowCredentials: false,
		}))
	}

	// Rate limiting middleware (if enabled)
	if s.config.RateLimit > 0 {
		s.echo.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
			Skipper: func(c echo.Context) bool {
				// Skip rate limiting for the random endpoint, as it only redirects.
				// Use c.Request().URL.Path to get the actual path, not the route pattern.
				return c.Request().URL.Path == "/api/v1/generate/random"
			},
			Store: middleware.NewRateLimiterMemoryStore(
				rate.Limit(float64(s.config.RateLimit) / 60.0), // Convert per-minute to per-second
			),
			IdentifierExtractor: func(c echo.Context) (string, error) {
				return c.RealIP(), nil
			},
			ErrorHandler: func(c echo.Context, err error) error {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error":   "rate limit identifier error",
					"message": "Could not identify the request source.",
				})
			},
			DenyHandler: func(c echo.Context, identifier string, err error) error {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error":   "rate limit exceeded",
					"message": "You have made too many requests. Please try again later.",
				})
			},
		}))
	}

	// Metrics middleware - tracks requests
	s.echo.Use(s.metricsMiddleware())

	// Request logging middleware (if enabled)
	if s.config.EnableLogging {
		s.echo.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogStatus:        true,
			LogURI:           true,
			LogMethod:        true,
			LogRemoteIP:      true,
			LogHost:          true,
			LogLatency:       true,
			LogError:         true,
			LogRequestID:     true,
			LogResponseSize:  true,
			LogContentLength: true,
			LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
				// Determine log level based on status code
				var level string
				switch {
				case v.Status >= 500:
					level = "error"
				case v.Status >= 400:
					level = "warn"
				default:
					level = "info"
				}

				// Skip logging if the level is below the configured threshold
				if !isLogLevelEnabled(level, s.config.LogLevel) {
					return nil
				}

				logEntry := map[string]interface{}{
					"time":          v.StartTime.Format(time.RFC3339Nano),
					"level":         level,
					"id":            v.RequestID,
					"remote_ip":     v.RemoteIP,
					"host":          v.Host,
					"method":        v.Method,
					"uri":           v.URI,
					"status":        v.Status,
					"latency_human": v.Latency.String(),
					"bytes_in":      v.ContentLength,
					"bytes_out":     v.ResponseSize,
				}

				if v.Error != nil {
					logEntry["error"] = v.Error.Error()
				}

				jsonBytes, err := json.Marshal(logEntry)
				if err != nil {
					log.Printf("Error marshaling log entry: %v", err)
					return nil
				}

				// Write directly to stdout
				os.Stdout.Write(jsonBytes)
				os.Stdout.Write([]byte("\n"))

				return nil
			},
		}))
	}

	// Recovery middleware
	s.echo.Use(middleware.Recover())

	// Body limit middleware to prevent huge requests
	s.echo.Use(middleware.BodyLimit(fmt.Sprintf("%d", s.config.MaxFileSize)))

	// Security headers middleware
	secureConfig := middleware.SecureConfig{
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      s.config.XFrameOptions,
	}

	// Apply stricter security settings in production
	if s.config.Environment == "production" {
		secureConfig.HSTSMaxAge = 31536000 // 1 year
		secureConfig.ContentSecurityPolicy = s.config.ContentSecurityPolicy
	}

	s.echo.Use(middleware.SecureWithConfig(secureConfig))
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.echo.GET("/health", handlers.NewHealthHandler(s.metrics))

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
	if len(s.config.CORSAllowedOrigins) > 0 {
		api.OPTIONS("/generate", func(c echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		})
		api.OPTIONS("/generate/:type", func(c echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		})
	}
}

// metricsMiddleware tracks request metrics and performance
func (s *Server) metricsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Record start time for performance tracking
			start := time.Now()

			// Increment request counter
			atomic.AddInt64(&s.metrics.RequestCount, 1)

			// Call the next handler
			err := next(c)

			// Calculate response time and record it
			duration := time.Since(start)
			s.metrics.Performance.AddResponseTime(duration)

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

// generateRequestID creates a secure request ID using timestamp and random string
func generateRequestID() string {
	// Generate cryptographically secure random bytes
	b := make([]byte, config.RequestIDLength)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return fmt.Sprintf("%d%s%d", time.Now().Unix(), config.RequestIDSeparator, time.Now().UnixNano()%1000)
	}

	// Create ID with timestamp and random component for uniqueness and readability
	return fmt.Sprintf("%d%s%x", time.Now().Unix(), config.RequestIDSeparator, b)
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
