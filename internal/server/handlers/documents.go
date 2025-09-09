package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/engineervix/pseudoc/internal/config"
)

// DocumentHandler handles document generation requests
type DocumentHandler struct {
	config *config.ServerConfig
}

// NewDocumentHandler creates a new document handler
func NewDocumentHandler(cfg *config.ServerConfig) *DocumentHandler {
	return &DocumentHandler{
		config: cfg,
	}
}

// GenerateGET handles GET requests for document generation
// URL format: /api/v1/generate/{type}?pages=3&sheets=2&seed=42
func (h *DocumentHandler) GenerateGET(c echo.Context) error {
	// TODO: Implement document generation
	// This is a placeholder for now

	docType := c.Param("type")

	return c.JSON(http.StatusNotImplemented, map[string]interface{}{
		"error":          "Not implemented yet",
		"message":        "Document generation is yet to be implemented",
		"requested_type": docType,
		"query_params":   c.QueryParams(),
	})
}

// GeneratePOST handles POST requests for document generation
// Expected JSON payload with document configuration
func (h *DocumentHandler) GeneratePOST(c echo.Context) error {
	// TODO: Implement document generation
	// This is a placeholder for now

	return c.JSON(http.StatusNotImplemented, map[string]interface{}{
		"error":   "Not implemented yet",
		"message": "Document generation is yet to be implemented",
		"info":    "Use POST with JSON body containing document configuration",
	})
}
