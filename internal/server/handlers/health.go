package handlers

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/engineervix/pseudoc/internal/version"
	"github.com/labstack/echo/v4"
)

// HealthResponse represents the health check response
// @Description Server health status information
type HealthResponse struct {
	Status    string        `json:"status" example:"healthy"`
	Timestamp time.Time     `json:"timestamp" example:"2024-01-15T10:30:00Z"`
	Version   string        `json:"version" example:"1.0.0"`
	Uptime    string        `json:"uptime" example:"2h30m15s"`
	System    *SystemInfo   `json:"system,omitempty"`
	Checks    []CheckResult `json:"checks,omitempty"`
}

// SystemInfo contains system-level metrics
type SystemInfo struct {
	MemoryUsageMB  uint64 `json:"memory_usage_mb"`
	GoroutineCount int    `json:"goroutine_count"`
}

// CheckResult represents the result of a single health check component
type CheckResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "ok", "warning", "error"
	Message string `json:"message,omitempty"`
}

// HealthChecker is an interface for components that can be health-checked
type HealthChecker interface {
	GetStartTime() time.Time
}

// NewHealthHandler creates a new health check handler
// @Summary Health check endpoint
// @Description Returns the current health status of the server
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse "Server is healthy"
// @Failure 503 {object} ErrorResponse "Server is unhealthy"
// @Router /health [get]
func NewHealthHandler(checker HealthChecker) echo.HandlerFunc {
	return func(c echo.Context) error {
		// System info
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		systemInfo := &SystemInfo{
			MemoryUsageMB:  memStats.Alloc / 1024 / 1024,
			GoroutineCount: runtime.NumGoroutine(),
		}

		// Uptime calculation
		uptime := time.Since(checker.GetStartTime())

		response := HealthResponse{
			Status:    "ok",
			Timestamp: time.Now(),
			Version:   version.Version,
			Uptime:    formatUptime(uptime),
			System:    systemInfo,
			Checks: []CheckResult{
				{
					Name:   "runtime",
					Status: "ok",
				},
			},
		}

		return c.JSON(http.StatusOK, response)
	}
}

// formatUptime formats duration into a human-readable string
func formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
}
