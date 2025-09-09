package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/engineervix/pseudoc/internal/config"
	"github.com/engineervix/pseudoc/internal/generator"
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

// GenerateRequest represents the JSON request body for POST /api/v1/generate
type GenerateRequest struct {
	Type     string `json:"type" validate:"required"`
	Pages    int    `json:"pages,omitempty"`
	Sheets   int    `json:"sheets,omitempty"`
	Seed     int64  `json:"seed,omitempty"`
	Filename string `json:"filename,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// GenerateGET handles GET requests for document generation
// URL format: /api/v1/generate/{type}?pages=3&sheets=2&seed=42
func (h *DocumentHandler) GenerateGET(c echo.Context) error {
	// Extract document type from URL parameter
	docType := strings.ToLower(c.Param("type"))

	// Create document config based on server defaults
	docConfig := h.config.Config

	// Parse query parameters
	if err := h.parseQueryParams(c, &docConfig, docType); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid query parameters",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	// Handle random document type selection
	if docType == "random" {
		var rng *rand.Rand
		if docConfig.Seed != 0 {
			rng = rand.New(rand.NewSource(docConfig.Seed))
		} else {
			rng = rand.New(rand.NewSource(time.Now().UnixNano()))
		}

		types := []string{config.DocTypePDF, config.DocTypeDOCX, config.DocTypeXLSX}
		docConfig.DocType = types[rng.Intn(len(types))]
	} else {
		// Validate and normalize document type
		normalizedType, err := h.normalizeDocType(docType)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid document type",
				Message: fmt.Sprintf("'%s' is not supported. Available types: pdf, docx, xlsx, random", docType),
				Code:    http.StatusBadRequest,
			})
		}
		docConfig.DocType = normalizedType
	}

	// Validate the configuration
	if err := docConfig.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid configuration",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	// Generate and stream the document
	return h.generateAndStream(c, &docConfig)
}

// GeneratePOST handles POST requests for document generation
// Expected JSON payload with document configuration
func (h *DocumentHandler) GeneratePOST(c echo.Context) error {
	var req GenerateRequest

	// Bind JSON request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid JSON payload",
			Message: "Failed to parse request body: " + err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	// Validate required fields
	if req.Type == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing required field",
			Message: "Field 'type' is required",
			Code:    http.StatusBadRequest,
		})
	}

	// Create document config from request
	docConfig := h.config.Config
	docConfig.Seed = req.Seed
	docConfig.BaseFilename = req.Filename

	// Set document-specific parameters
	if req.Pages > 0 {
		docConfig.Pages = req.Pages
	}
	if req.Sheets > 0 {
		docConfig.Sheets = req.Sheets
	}

	// Handle random document type selection
	if strings.ToLower(req.Type) == "random" {
		var rng *rand.Rand
		if docConfig.Seed != 0 {
			rng = rand.New(rand.NewSource(docConfig.Seed))
		} else {
			rng = rand.New(rand.NewSource(time.Now().UnixNano()))
		}

		types := []string{config.DocTypePDF, config.DocTypeDOCX, config.DocTypeXLSX}
		docConfig.DocType = types[rng.Intn(len(types))]
	} else {
		// Validate and normalize document type
		normalizedType, err := h.normalizeDocType(req.Type)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid document type",
				Message: fmt.Sprintf("'%s' is not supported. Available types: pdf, docx, xlsx, random", req.Type),
				Code:    http.StatusBadRequest,
			})
		}
		docConfig.DocType = normalizedType
	}

	// Validate the configuration
	if err := docConfig.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid configuration",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	// Generate and stream the document
	return h.generateAndStream(c, &docConfig)
}

// parseQueryParams parses query parameters and updates the document config
func (h *DocumentHandler) parseQueryParams(c echo.Context, cfg *config.Config, docType string) error {
	// Parse pages parameter
	if pagesStr := c.QueryParam("pages"); pagesStr != "" {
		pages, err := strconv.Atoi(pagesStr)
		if err != nil || pages < 1 {
			return fmt.Errorf("pages must be a positive integer, got '%s'", pagesStr)
		}
		if pages > 100 { // Reasonable limit
			return fmt.Errorf("pages cannot exceed 100, got %d", pages)
		}
		cfg.Pages = pages
	}

	// Parse sheets parameter
	if sheetsStr := c.QueryParam("sheets"); sheetsStr != "" {
		sheets, err := strconv.Atoi(sheetsStr)
		if err != nil || sheets < 1 {
			return fmt.Errorf("sheets must be a positive integer, got '%s'", sheetsStr)
		}
		if sheets > 50 { // Reasonable limit
			return fmt.Errorf("sheets cannot exceed 50, got %d", sheets)
		}
		cfg.Sheets = sheets
	}

	// Parse seed parameter
	if seedStr := c.QueryParam("seed"); seedStr != "" {
		seed, err := strconv.ParseInt(seedStr, 10, 64)
		if err != nil {
			return fmt.Errorf("seed must be a valid integer, got '%s'", seedStr)
		}
		cfg.Seed = seed
	}

	// Parse filename parameter
	if filename := c.QueryParam("filename"); filename != "" {
		// Basic validation for filename
		if strings.ContainsAny(filename, "/\\:*?\"<>|") {
			return fmt.Errorf("filename contains invalid characters")
		}
		cfg.BaseFilename = filename
	}

	return nil
}

// normalizeDocType converts various document type aliases to standard types
func (h *DocumentHandler) normalizeDocType(docType string) (string, error) {
	switch strings.ToLower(docType) {
	case "pdf":
		return config.DocTypePDF, nil
	case "docx", "word":
		return config.DocTypeDOCX, nil
	case "xlsx", "spreadsheet", "sheet":
		return config.DocTypeXLSX, nil
	default:
		return "", fmt.Errorf("unsupported document type: %s", docType)
	}
}

// generateAndStream generates the document and streams it to the client
func (h *DocumentHandler) generateAndStream(c echo.Context, cfg *config.Config) error {
	// Get the appropriate generator
	gen, err := generator.GetGenerator(cfg.DocType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Generator error",
			Message: "Failed to get document generator: " + err.Error(),
			Code:    http.StatusInternalServerError,
		})
	}

	// Generate filename for response headers
	filename := h.generateFilename(cfg)

	// Set response headers
	c.Response().Header().Set("Content-Type", h.getContentType(cfg.DocType))
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Set cache headers to prevent caching of generated documents
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")

	// Generate the document and stream it directly to the response
	c.Response().WriteHeader(http.StatusOK)

	if err := gen.Generate(c.Response().Writer, cfg); err != nil {
		// At this point headers are already sent, so we can't return a JSON error
		// Log the error and close the connection
		c.Logger().Errorf("Document generation failed: %v", err)
		return err
	}

	return nil
}

// generateFilename creates a filename for the response
func (h *DocumentHandler) generateFilename(cfg *config.Config) string {
	base := cfg.BaseFilename
	if base == "" {
		base = "pseudoc_" + time.Now().Format("2006-01-02_15-04-05")
	}

	return base + cfg.GetFileExtension()
}

// getContentType returns the appropriate Content-Type header for the document type
func (h *DocumentHandler) getContentType(docType string) string {
	switch docType {
	case config.DocTypePDF:
		return "application/pdf"
	case config.DocTypeDOCX:
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case config.DocTypeXLSX:
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}
