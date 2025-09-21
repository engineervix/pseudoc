package handlers

import (
	"bytes"
	"fmt"
	"io"
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
	config  *config.ServerConfig
	metrics ServerMetricsInterface
}

// ServerMetricsInterface defines methods for updating metrics
type ServerMetricsInterface interface {
	IncrementDocumentCount(docType string)
}

// NewDocumentHandler creates a new document handler with metrics support
func NewDocumentHandler(cfg *config.ServerConfig, metrics ServerMetricsInterface) *DocumentHandler {
	return &DocumentHandler{
		config:  cfg,
		metrics: metrics,
	}
}

// GenerateRequest represents the JSON request body for POST /api/v1/generate
// @Description Request body for document generation via POST
type GenerateRequest struct {
	Type     string `json:"type" example:"pdf" enums:"pdf,docx,xlsx,random" validate:"required"`
	Pages    int    `json:"pages,omitempty" example:"5" minimum:"1" maximum:"100"`
	Sheets   int    `json:"sheets,omitempty" example:"3" minimum:"1" maximum:"10"`
	Seed     int64  `json:"seed,omitempty" example:"42"`
	Filename string `json:"filename,omitempty" example:"my-document"`
}

// ValidationError represents a validation error with details
// @Description Detailed validation error information
type ValidationError struct {
	Field   string `json:"field" example:"pages"`
	Value   string `json:"value" example:"150"`
	Message string `json:"message" example:"pages must be between 1 and 100"`
}

// ErrorResponse represents an error response with enhanced information
// @Description Standard error response format
type ErrorResponse struct {
	Error      string            `json:"error" example:"Invalid document type"`
	Message    string            `json:"message" example:"Supported types: pdf, docx, xlsx, random"`
	Code       int               `json:"code" example:"400"`
	RequestID  string            `json:"request_id,omitempty" example:"req_123456789"`
	Validation []ValidationError `json:"validation_errors,omitempty"`
}

// GenerateGET handles GET requests for document generation
// @Summary Generate document via GET request
// @Description Generate a document using URL path parameter and query parameters. Supports PDF, DOCX, XLSX, and random document types.
// @Tags documents
// @Accept json
// @Produce application/pdf,application/vnd.openxmlformats-officedocument.wordprocessingml.document,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param type path string true "Document type" Enums(pdf,docx,xlsx,random) example(pdf)
// @Param pages query int false "Number of pages (PDF/DOCX only)" minimum(1) maximum(100) default(3) example(5)
// @Param sheets query int false "Number of sheets (XLSX only)" minimum(1) maximum(10) default(2) example(3)
// @Param seed query int64 false "Random seed for reproducible generation" example(42)
// @Param filename query string false "Custom filename (without extension)" example(my-document)
// @Success 200 {file} file "Generated document file"
// @Failure 400 {object} ErrorResponse "Invalid parameters"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /generate/{type} [get]
func (h *DocumentHandler) GenerateGET(c echo.Context) error {
	// Get request ID for error responses
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	// Extract document type from URL parameter
	docType := strings.ToLower(c.Param("type"))

	// Create document config based on server defaults
	docConfig := h.config.Config

	// Parse and validate query parameters
	if err := h.parseAndValidateQueryParams(c, &docConfig, docType); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:     "Invalid query parameters",
			Message:   err.Error(),
			Code:      http.StatusBadRequest,
			RequestID: requestID,
		})
	}

	// Handle random document type selection or validate type
	if err := h.processDocumentType(docType, &docConfig); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:     "Invalid document type",
			Message:   err.Error(),
			Code:      http.StatusBadRequest,
			RequestID: requestID,
		})
	}

	// Validate the final configuration
	if err := h.validateDocumentConfig(&docConfig); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:     "Invalid configuration",
			Message:   err.Error(),
			Code:      http.StatusBadRequest,
			RequestID: requestID,
		})
	}

	// Generate and stream the document
	return h.generateAndStream(c, &docConfig)
}

// GeneratePOST handles POST requests for document generation
// @Summary Generate document via POST request
// @Description Generate a document using JSON request body. Provides more flexibility than GET endpoint.
// @Tags documents
// @Accept json
// @Produce application/pdf,application/vnd.openxmlformats-officedocument.wordprocessingml.document,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param request body GenerateRequest true "Document generation parameters"
// @Success 200 {file} file "Generated document file"
// @Failure 400 {object} ErrorResponse "Invalid request body or parameters"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /generate [post]
func (h *DocumentHandler) GeneratePOST(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	var req GenerateRequest

	// Bind JSON request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:     "Invalid JSON payload",
			Message:   "Failed to parse request body: " + err.Error(),
			Code:      http.StatusBadRequest,
			RequestID: requestID,
		})
	}

	// Validate the request
	if validationErrors := h.validateGenerateRequest(&req); len(validationErrors) > 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:      "Validation failed",
			Message:    "One or more fields contain invalid values",
			Code:       http.StatusBadRequest,
			RequestID:  requestID,
			Validation: validationErrors,
		})
	}

	// Create document config from request
	docConfig := h.buildConfigFromRequest(&req)

	// Handle random document type selection or validate type
	if err := h.processDocumentType(req.Type, &docConfig); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:     "Invalid document type",
			Message:   err.Error(),
			Code:      http.StatusBadRequest,
			RequestID: requestID,
		})
	}

	// Validate the final configuration
	if err := h.validateDocumentConfig(&docConfig); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:     "Invalid configuration",
			Message:   err.Error(),
			Code:      http.StatusBadRequest,
			RequestID: requestID,
		})
	}

	// Generate and stream the document
	return h.generateAndStream(c, &docConfig)
}

// parseAndValidateQueryParams parses and validates query parameters
func (h *DocumentHandler) parseAndValidateQueryParams(c echo.Context, cfg *config.Config, docType string) error {
	// Parse pages parameter
	if pagesStr := c.QueryParam("pages"); pagesStr != "" {
		pages, err := strconv.Atoi(pagesStr)
		if err != nil {
			return fmt.Errorf("pages parameter must be a valid integer, got '%s'", pagesStr)
		}
		if pages < 1 {
			return fmt.Errorf("pages must be at least 1, got %d", pages)
		}
		if pages > config.MaxPages {
			return fmt.Errorf("pages cannot exceed %d, got %d", config.MaxPages, pages)
		}
		cfg.Pages = pages
	}

	// Parse sheets parameter
	if sheetsStr := c.QueryParam("sheets"); sheetsStr != "" {
		sheets, err := strconv.Atoi(sheetsStr)
		if err != nil {
			return fmt.Errorf("sheets parameter must be a valid integer, got '%s'", sheetsStr)
		}
		if sheets < 1 {
			return fmt.Errorf("sheets must be at least 1, got %d", sheets)
		}
		if sheets > config.MaxSheets {
			return fmt.Errorf("sheets cannot exceed %d, got %d", config.MaxSheets, sheets)
		}
		cfg.Sheets = sheets
	}

	// Parse seed parameter
	if seedStr := c.QueryParam("seed"); seedStr != "" {
		seed, err := strconv.ParseInt(seedStr, 10, 64)
		if err != nil {
			return fmt.Errorf("seed parameter must be a valid integer, got '%s'", seedStr)
		}
		cfg.Seed = seed
	}

	// Parse and validate filename parameter
	if filename := c.QueryParam("filename"); filename != "" {
		if err := h.validateFilename(filename); err != nil {
			return fmt.Errorf("filename parameter: %w", err)
		}
		cfg.BaseFilename = filename
	}

	return nil
}

// validateGenerateRequest validates the POST request payload
func (h *DocumentHandler) validateGenerateRequest(req *GenerateRequest) []ValidationError {
	var errors []ValidationError

	// Required field validation
	if req.Type == "" {
		errors = append(errors, ValidationError{
			Field:   "type",
			Value:   req.Type,
			Message: "type is required and cannot be empty",
		})
	}

	// Pages validation
	if req.Pages < 0 {
		errors = append(errors, ValidationError{
			Field:   "pages",
			Value:   fmt.Sprintf("%d", req.Pages),
			Message: "pages cannot be negative",
		})
	}
	if req.Pages > config.MaxPages {
		errors = append(errors, ValidationError{
			Field:   "pages",
			Value:   fmt.Sprintf("%d", req.Pages),
			Message: fmt.Sprintf("pages cannot exceed %d", config.MaxPages),
		})
	}

	// Sheets validation
	if req.Sheets < 0 {
		errors = append(errors, ValidationError{
			Field:   "sheets",
			Value:   fmt.Sprintf("%d", req.Sheets),
			Message: "sheets cannot be negative",
		})
	}
	if req.Sheets > config.MaxSheets {
		errors = append(errors, ValidationError{
			Field:   "sheets",
			Value:   fmt.Sprintf("%d", req.Sheets),
			Message: fmt.Sprintf("sheets cannot exceed %d", config.MaxSheets),
		})
	}

	// Filename validation
	if req.Filename != "" {
		if err := h.validateFilename(req.Filename); err != nil {
			errors = append(errors, ValidationError{
				Field:   "filename",
				Value:   req.Filename,
				Message: err.Error(),
			})
		}
	}

	return errors
}

// validateFilename validates a filename for security and compatibility
func (h *DocumentHandler) validateFilename(filename string) error {
	if len(filename) == 0 {
		return fmt.Errorf("filename cannot be empty")
	}
	if len(filename) > config.MaxFilenameLength {
		return fmt.Errorf("filename too long (max %d characters)", config.MaxFilenameLength)
	}

	// Check for invalid characters (Windows + Unix)
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\n", "\r", "\t"}
	for _, char := range invalidChars {
		if strings.Contains(filename, char) {
			return fmt.Errorf("filename contains invalid character '%s'", char)
		}
	}

	// Check for reserved names (Windows)
	upperFilename := strings.ToUpper(filename)
	for _, name := range config.WindowsReservedFilenames {
		if upperFilename == name {
			return fmt.Errorf("filename '%s' is reserved", filename)
		}
	}

	return nil
}

// buildConfigFromRequest creates a document config from the request
func (h *DocumentHandler) buildConfigFromRequest(req *GenerateRequest) config.Config {
	docConfig := h.config.Config
	docConfig.Seed = req.Seed
	docConfig.BaseFilename = req.Filename

	// Set document-specific parameters with defaults
	if req.Pages > 0 {
		docConfig.Pages = req.Pages
	}
	if req.Sheets > 0 {
		docConfig.Sheets = req.Sheets
	}

	return docConfig
}

// processDocumentType handles random document type selection or validates the type
func (h *DocumentHandler) processDocumentType(docType string, cfg *config.Config) error {
	if strings.ToLower(docType) == "random" {
		var rng *rand.Rand
		if cfg.Seed != 0 {
			rng = rand.New(rand.NewSource(cfg.Seed))
		} else {
			rng = rand.New(rand.NewSource(time.Now().UnixNano()))
		}

		types := []string{config.DocTypePDF, config.DocTypeDOCX, config.DocTypeXLSX}
		cfg.DocType = types[rng.Intn(len(types))]
		return nil
	}

	// Validate and normalize document type
	normalizedType, err := h.normalizeDocType(docType)
	if err != nil {
		return fmt.Errorf("'%s' is not supported. Available types: pdf, docx, xlsx, random", docType)
	}
	cfg.DocType = normalizedType
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

// validateDocumentConfig validates the final document configuration
func (h *DocumentHandler) validateDocumentConfig(cfg *config.Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}
	return nil
}

// generateAndStream generates the document and streams it to the client
func (h *DocumentHandler) generateAndStream(c echo.Context, cfg *config.Config) error {
	// Get the appropriate generator
	gen, err := generator.GetGenerator(cfg.DocType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:     "Generator error",
			Message:   "Failed to get document generator: " + err.Error(),
			Code:      http.StatusInternalServerError,
			RequestID: c.Response().Header().Get(echo.HeaderXRequestID),
		})
	}

	// Generate filename for response headers
	filename := h.generateFilename(cfg)

	// Generate the document to a buffer first to handle errors properly
	var buf bytes.Buffer
	if err := gen.Generate(&buf, cfg); err != nil {
		// Since we haven't sent headers yet, we can return a proper JSON error
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:     "Document generation failed",
			Message:   "Failed to generate document: " + err.Error(),
			Code:      http.StatusInternalServerError,
			RequestID: c.Response().Header().Get(echo.HeaderXRequestID),
		})
	}

	// Check if the generated document exceeds the maximum file size
	if int64(buf.Len()) > h.config.MaxFileSize {
		return c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{
			Error:     "Document too large",
			Message:   fmt.Sprintf("Generated document size (%d bytes) exceeds maximum allowed size (%d bytes)", buf.Len(), h.config.MaxFileSize),
			Code:      http.StatusRequestEntityTooLarge,
			RequestID: c.Response().Header().Get(echo.HeaderXRequestID),
		})
	}

	// Update metrics only after successful generation
	if h.metrics != nil {
		h.metrics.IncrementDocumentCount(cfg.DocType)
	}

	// Set response headers
	c.Response().Header().Set("Content-Type", h.getContentType(cfg.DocType))
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Response().Header().Set("X-Pseudoc-Generated-Type", cfg.DocType)
	c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))

	// Set cache headers to prevent caching of generated documents
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")

	// Stream the buffer content to the response
	c.Response().WriteHeader(http.StatusOK)
	_, err = io.Copy(c.Response().Writer, &buf)
	if err != nil {
		// Log the error, but we can't change the response at this point
		c.Logger().Errorf("Failed to stream document for request %s: %v",
			c.Response().Header().Get(echo.HeaderXRequestID), err)
		return err
	}

	return nil
}

// generateFilename creates a filename for the response
func (h *DocumentHandler) generateFilename(cfg *config.Config) string {
	base := cfg.BaseFilename
	if base == "" {
		base = config.DefaultFilenamePrefix + time.Now().Format(config.DefaultTimestampFormat)
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
