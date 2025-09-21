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
type HealthResponse struct {
	Status    string        `json:"status"`
	Timestamp time.Time     `json:"timestamp"`
	Version   string        `json:"version"`
	Uptime    string        `json:"uptime"`
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
