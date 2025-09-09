package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/engineervix/pseudoc/internal/config"
)

// InfoResponse represents server information
type InfoResponse struct {
	Name             string            `json:"name"`
	Version          string            `json:"version"`
	Description      string            `json:"description"`
	SupportedFormats []string          `json:"supported_formats"`
	Endpoints        map[string]string `json:"endpoints"`
	ServerInfo       ServerInfo        `json:"server_info"`
	Limits           Limits            `json:"limits"`
}

// ServerInfo contains runtime information
type ServerInfo struct {
	GoVersion    string    `json:"go_version"`
	Platform     string    `json:"platform"`
	Architecture string    `json:"architecture"`
	StartTime    time.Time `json:"start_time"`
	Uptime       string    `json:"uptime"`
}

// Limits contains server limits and configuration
type Limits struct {
	MaxFileSize    int64         `json:"max_file_size_bytes"`
	RequestTimeout time.Duration `json:"request_timeout"`
	RateLimit      int           `json:"rate_limit_per_minute"`
	MaxPages       int           `json:"max_pages"`
	MaxSheets      int           `json:"max_sheets"`
}

var serverStartTime = time.Now()

// NewInfoHandler creates a new server info handler
func NewInfoHandler(cfg *config.ServerConfig) echo.HandlerFunc {
	return func(c echo.Context) error {
		baseURL := cfg.GetBaseURL()

		response := InfoResponse{
			Name:        "pseudoc",
			Version:     "0.1.0", // TODO: Get this from build info
			Description: "Lorem Ipsum for Documents - Generate placeholder documents for testing and development",
			SupportedFormats: []string{
				config.DocTypePDF,
				config.DocTypeDOCX,
				config.DocTypeXLSX,
				"random",
			},
			Endpoints: map[string]string{
				"health":        baseURL + "/health",
				"info":          baseURL + "/api/v1/info",
				"generate_get":  baseURL + "/api/v1/generate/{type}?pages=3&sheets=2&seed=42",
				"generate_post": baseURL + "/api/v1/generate",
				"examples":      baseURL + "/api/v1/generate/pdf?pages=3",
			},
			ServerInfo: ServerInfo{
				GoVersion:    runtime.Version(),
				Platform:     runtime.GOOS,
				Architecture: runtime.GOARCH,
				StartTime:    serverStartTime,
				Uptime:       time.Since(serverStartTime).String(),
			},
			Limits: Limits{
				MaxFileSize:    cfg.MaxFileSize,
				RequestTimeout: cfg.RequestTimeout,
				RateLimit:      cfg.RateLimit,
				MaxPages:       100, // Reasonable defaults
				MaxSheets:      50,
			},
		}

		return c.JSON(http.StatusOK, response)
	}
}
