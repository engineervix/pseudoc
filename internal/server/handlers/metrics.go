package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
)

// MetricsResponse represents basic server metrics
type MetricsResponse struct {
	Uptime          string           `json:"uptime"`
	MemoryUsage     MemoryStats      `json:"memory_usage"`
	GoroutineCount  int              `json:"goroutine_count"`
	RequestCount    int64            `json:"request_count,omitempty"`    // TODO: Implement counter
	GenerationCount map[string]int64 `json:"generation_count,omitempty"` // TODO: Implement counters
}

// MemoryStats contains memory usage information
type MemoryStats struct {
	AllocMB      uint64 `json:"allocated_mb"`
	TotalAllocMB uint64 `json:"total_allocated_mb"`
	SystemMB     uint64 `json:"system_mb"`
	GCCount      uint32 `json:"gc_count"`
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		response := MetricsResponse{
			Uptime:         time.Since(serverStartTime).String(),
			GoroutineCount: runtime.NumGoroutine(),
			MemoryUsage: MemoryStats{
				AllocMB:      bToMb(m.Alloc),
				TotalAllocMB: bToMb(m.TotalAlloc),
				SystemMB:     bToMb(m.Sys),
				GCCount:      m.NumGC,
			},
			// TODO: Add request counters and generation counters
			// These would be implemented with sync/atomic counters
		}

		return c.JSON(http.StatusOK, response)
	}
}

// bToMb converts bytes to megabytes
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
