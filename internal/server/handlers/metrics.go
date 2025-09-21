package handlers

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/engineervix/pseudoc/internal/config"
)

// MetricsResponse represents comprehensive server metrics
type MetricsResponse struct {
	Server      ServerMetrics      `json:"server"`
	Runtime     RuntimeMetrics     `json:"runtime"`
	Documents   DocumentMetrics    `json:"documents"`
	Performance PerformanceMetrics `json:"performance"`
}

// ServerMetrics contains server-level statistics
type ServerMetrics struct {
	Uptime       string `json:"uptime"`
	StartTime    string `json:"start_time"`
	RequestCount int64  `json:"total_requests"`
	ErrorCount   int64  `json:"total_errors"`
	ErrorRate    string `json:"error_rate"`
}

// RuntimeMetrics contains Go runtime information
type RuntimeMetrics struct {
	GoVersion      string   `json:"go_version"`
	GoroutineCount int      `json:"goroutine_count"`
	Memory         MemStats `json:"memory"`
	GC             GCStats  `json:"garbage_collection"`
}

// DocumentMetrics contains document generation statistics
type DocumentMetrics struct {
	TotalGenerated int64            `json:"total_generated"`
	ByType         map[string]int64 `json:"by_type"`
}

// PerformanceMetrics contains performance-related metrics
type PerformanceMetrics struct {
	AverageResponseTime string `json:"avg_response_time,omitempty"`
	RequestsPerMinute   string `json:"requests_per_minute,omitempty"`
}

// MemStats contains detailed memory usage information
type MemStats struct {
	AllocatedMB  uint64 `json:"allocated_mb"`
	TotalAllocMB uint64 `json:"total_allocated_mb"`
	SystemMB     uint64 `json:"system_mb"`
	HeapInUseMB  uint64 `json:"heap_inuse_mb"`
	HeapIdleMB   uint64 `json:"heap_idle_mb"`
	StackInUseMB uint64 `json:"stack_inuse_mb"`
}

// GCStats contains garbage collection statistics
type GCStats struct {
	NumGC        uint32 `json:"num_gc"`
	LastGC       string `json:"last_gc"`
	PauseTotalMs uint64 `json:"pause_total_ms"`
	NextGCMB     uint64 `json:"next_gc_mb"`
}

// MetricsProvider interface defines what the metrics handler needs
type MetricsProvider interface {
	GetRequestCount() int64
	GetErrorCount() int64
	GetDocumentCounts() map[string]int64
	GetStartTime() time.Time
	GetAverageResponseTime() time.Duration
	GetRequestsPerMinute() float64
	HasSufficientSamples() bool
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(metricsProvider MetricsProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// Get metrics from provider
		requestCount := metricsProvider.GetRequestCount()
		errorCount := metricsProvider.GetErrorCount()
		docCounts := metricsProvider.GetDocumentCounts()
		startTime := metricsProvider.GetStartTime()

		// Calculate total documents generated
		totalDocs := int64(0)
		for _, count := range docCounts {
			totalDocs += count
		}

		// Calculate error rate
		errorRate := "0.0%"
		if requestCount > 0 {
			rate := float64(errorCount) / float64(requestCount) * 100
			errorRate = fmt.Sprintf("%.1f%%", rate)
		}

		// Calculate uptime
		uptime := time.Since(startTime)

		// Format last GC time
		lastGC := "never"
		if m.LastGC > 0 {
			lastGC = time.Unix(0, int64(m.LastGC)).Format(time.RFC3339)
		}

		// Get performance metrics
		avgResponseTime := "insufficient_data"
		requestsPerMinute := "insufficient_data"

		if metricsProvider.HasSufficientSamples() {
			if avgDuration := metricsProvider.GetAverageResponseTime(); avgDuration > 0 {
				avgResponseTime = avgDuration.String()
			}

			rpm := metricsProvider.GetRequestsPerMinute()
			requestsPerMinute = fmt.Sprintf("%.1f", rpm)
		}

		response := MetricsResponse{
			Server: ServerMetrics{
				Uptime:       uptime.String(),
				StartTime:    startTime.Format(time.RFC3339),
				RequestCount: requestCount,
				ErrorCount:   errorCount,
				ErrorRate:    errorRate,
			},
			Runtime: RuntimeMetrics{
				GoVersion:      runtime.Version(),
				GoroutineCount: runtime.NumGoroutine(),
				Memory: MemStats{
					AllocatedMB:  bToMb(m.Alloc),
					TotalAllocMB: bToMb(m.TotalAlloc),
					SystemMB:     bToMb(m.Sys),
					HeapInUseMB:  bToMb(m.HeapInuse),
					HeapIdleMB:   bToMb(m.HeapIdle),
					StackInUseMB: bToMb(m.StackInuse),
				},
				GC: GCStats{
					NumGC:        m.NumGC,
					LastGC:       lastGC,
					PauseTotalMs: m.PauseTotalNs / config.NanosecondsToMilliseconds, // Convert to milliseconds
					NextGCMB:     bToMb(m.NextGC),
				},
			},
			Documents: DocumentMetrics{
				TotalGenerated: totalDocs,
				ByType:         docCounts,
			},
			Performance: PerformanceMetrics{
				AverageResponseTime: avgResponseTime,
				RequestsPerMinute:   requestsPerMinute,
			},
		}

		return c.JSON(http.StatusOK, response)
	}
}

// bToMb converts bytes to megabytes
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
