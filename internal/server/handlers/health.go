package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version,omitempty"`
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		response := HealthResponse{
			Status:    "ok",
			Timestamp: time.Now(),
			Version:   "0.1.0", // TODO: Get this from build info
		}

		return c.JSON(http.StatusOK, response)
	}
}
