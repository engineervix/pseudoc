package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/engineervix/pseudoc/internal/config"
)

// Mock metrics implementation for testing
type mockMetrics struct {
	documentCounts map[string]int64
}

func (m *mockMetrics) IncrementDocumentCount(docType string) {
	if m.documentCounts == nil {
		m.documentCounts = make(map[string]int64)
	}
	m.documentCounts[docType]++
}

func newMockMetrics() *mockMetrics {
	return &mockMetrics{
		documentCounts: make(map[string]int64),
	}
}

func TestDocumentHandler_GenerateGET_PDF(t *testing.T) {
	// Setup
	cfg := config.DefaultServerConfig()
	handler := NewDocumentHandler(&cfg, newMockMetrics())

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/generate/pdf", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("type")
	c.SetParamValues("pdf")

	// Execute
	err := handler.GenerateGET(c)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Check Content-Type header
	expectedContentType := "application/pdf"
	if contentType := rec.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
	}

	// Check Content-Disposition header
	contentDisposition := rec.Header().Get("Content-Disposition")
	if !strings.Contains(contentDisposition, "attachment") || !strings.Contains(contentDisposition, ".pdf") {
		t.Errorf("Expected Content-Disposition to contain 'attachment' and '.pdf', got %s", contentDisposition)
	}

	// Check that we got PDF content (starts with %PDF)
	body := rec.Body.Bytes()
	if len(body) < 4 || string(body[:4]) != "%PDF" {
		t.Errorf("Expected PDF content, got %d bytes", len(body))
	}
}

func TestDocumentHandler_GenerateGET_DOCX(t *testing.T) {
	cfg := config.DefaultServerConfig()
	handler := NewDocumentHandler(&cfg, newMockMetrics())

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/generate/docx", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("type")
	c.SetParamValues("docx")

	err := handler.GenerateGET(c)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	expectedContentType := "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	if contentType := rec.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
	}

	// Check for ZIP signature (DOCX files are ZIP archives)
	body := rec.Body.Bytes()
	if len(body) < 2 || body[0] != 0x50 || body[1] != 0x4B {
		t.Errorf("Expected DOCX content (ZIP signature), got %d bytes", len(body))
	}
}

func TestDocumentHandler_GenerateGET_XLSX(t *testing.T) {
	cfg := config.DefaultServerConfig()
	handler := NewDocumentHandler(&cfg, newMockMetrics())

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/generate/xlsx", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("type")
	c.SetParamValues("xlsx")

	err := handler.GenerateGET(c)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	expectedContentType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	if contentType := rec.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
	}
}

func TestDocumentHandler_GenerateGET_WithQueryParams(t *testing.T) {
	cfg := config.DefaultServerConfig()
	handler := NewDocumentHandler(&cfg, newMockMetrics())

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/generate/pdf?pages=3&seed=42&filename=test-doc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("type")
	c.SetParamValues("pdf")

	err := handler.GenerateGET(c)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Check that custom filename is used
	contentDisposition := rec.Header().Get("Content-Disposition")
	if !strings.Contains(contentDisposition, "test-doc.pdf") {
		t.Errorf("Expected filename 'test-doc.pdf' in Content-Disposition, got %s", contentDisposition)
	}
}

func TestDocumentHandler_GenerateGET_Random(t *testing.T) {
	cfg := config.DefaultServerConfig()
	handler := NewDocumentHandler(&cfg, newMockMetrics())

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/generate/random?seed=42", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("type")
	c.SetParamValues("random")

	err := handler.GenerateGET(c)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if rec.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status 307, got %d", rec.Code)
	}

	// With a fixed seed, the redirect location should be deterministic
	location := rec.Header().Get("Location")
	if !strings.HasPrefix(location, "/api/v1/generate/") {
		t.Errorf("Expected redirect to a generate endpoint, got %s", location)
	}
	if !strings.Contains(location, "seed=42") {
		t.Errorf("Expected redirect URL to contain the original query parameters, got %s", location)
	}

	// Verify that the redirected document type is one of the valid types
	validTypes := []string{"pdf", "docx", "xlsx"}
	isTypeValid := false
	for _, validType := range validTypes {
		if strings.Contains(location, "/"+validType+"?") {
			isTypeValid = true
			break
		}
	}
	if !isTypeValid {
		t.Errorf("Redirect URL does not contain a valid document type: %s", location)
	}
}

func TestDocumentHandler_GenerateGET_InvalidType(t *testing.T) {
	cfg := config.DefaultServerConfig()
	handler := NewDocumentHandler(&cfg, newMockMetrics())

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/generate/invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("type")
	c.SetParamValues("invalid")

	err := handler.GenerateGET(c)

	if err != nil {
		t.Fatalf("Expected no error (should return JSON error), got: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}

	var errorResp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errorResp); err != nil {
		t.Fatalf("Failed to parse error response: %v", err)
	}

	if errorResp.Error != "Invalid document type" {
		t.Errorf("Expected 'Invalid document type', got %s", errorResp.Error)
	}
}

func TestDocumentHandler_GenerateGET_InvalidQueryParams(t *testing.T) {
	cfg := config.DefaultServerConfig()
	handler := NewDocumentHandler(&cfg, newMockMetrics())

	testCases := []struct {
		name  string
		query string
	}{
		{"invalid pages", "pages=invalid"},
		{"negative pages", "pages=-1"},
		{"too many pages", "pages=200"},
		{"invalid sheets", "sheets=invalid"},
		{"invalid seed", "seed=invalid"},
		{"invalid filename", "filename=test/file"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/generate/pdf?"+tc.query, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("type")
			c.SetParamValues("pdf")

			err := handler.GenerateGET(c)

			if err != nil {
				t.Fatalf("Expected no error (should return JSON error), got: %v", err)
			}

			if rec.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got %d", rec.Code)
			}
		})
	}
}

func TestDocumentHandler_GeneratePOST_Success(t *testing.T) {
	cfg := config.DefaultServerConfig()
	handler := NewDocumentHandler(&cfg, newMockMetrics())

	requestBody := GenerateRequest{
		Type:     "pdf",
		Pages:    3,
		Seed:     42,
		Filename: "test-document",
	}

	jsonBody, _ := json.Marshal(requestBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/generate", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GeneratePOST(c)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Check that custom filename is used
	contentDisposition := rec.Header().Get("Content-Disposition")
	if !strings.Contains(contentDisposition, "test-document.pdf") {
		t.Errorf("Expected filename 'test-document.pdf' in Content-Disposition, got %s", contentDisposition)
	}
}

func TestDocumentHandler_GeneratePOST_Random(t *testing.T) {
	cfg := config.DefaultServerConfig()
	handler := NewDocumentHandler(&cfg, newMockMetrics())

	requestBody := GenerateRequest{
		Type: "random",
		Seed: 42,
	}

	jsonBody, _ := json.Marshal(requestBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/generate", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GeneratePOST(c)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestDocumentHandler_GeneratePOST_InvalidJSON(t *testing.T) {
	cfg := config.DefaultServerConfig()
	handler := NewDocumentHandler(&cfg, newMockMetrics())

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/generate", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GeneratePOST(c)

	if err != nil {
		t.Fatalf("Expected no error (should return JSON error), got: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestDocumentHandler_GenerateAndStream_ErrorHandling(t *testing.T) {
	// Setup with very small max file size to trigger size limit
	cfg := config.DefaultServerConfig()
	cfg.MaxFileSize = 100 // Very small limit to trigger error
	handler := NewDocumentHandler(&cfg, newMockMetrics())

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/generate/pdf", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("type")
	c.SetParamValues("pdf")

	// Execute
	err := handler.GenerateGET(c)

	// Assert - should get a proper JSON error response, not a streaming error
	if err != nil {
		t.Errorf("Expected no error from handler, got %v", err)
	}

	// Should get 413 Request Entity Too Large
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status 413 (Request Entity Too Large), got %d", rec.Code)
	}

	// Should get proper JSON error response
	var errorResp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errorResp); err != nil {
		t.Errorf("Failed to parse error response: %v", err)
	}

	if errorResp.Error != "Document too large" {
		t.Errorf("Expected error 'Document too large', got '%s'", errorResp.Error)
	}

	if errorResp.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected error code 413, got %d", errorResp.Code)
	}
}

func TestDocumentHandler_GeneratePOST_MissingType(t *testing.T) {
	cfg := config.DefaultServerConfig()
	handler := NewDocumentHandler(&cfg, newMockMetrics())

	requestBody := GenerateRequest{
		Pages: 3,
	}

	jsonBody, _ := json.Marshal(requestBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/generate", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GeneratePOST(c)

	if err != nil {
		t.Fatalf("Expected no error (should return JSON error), got: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}

	var errorResp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errorResp); err != nil {
		t.Fatalf("Failed to parse error response: %v", err)
	}

	if errorResp.Error != "Validation failed" {
		t.Errorf("Expected 'Validation failed', got %s", errorResp.Error)
	}
}

func TestDocumentHandler_NormalizeDocType(t *testing.T) {
	handler := &DocumentHandler{}

	testCases := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"pdf", config.DocTypePDF, false},
		{"PDF", config.DocTypePDF, false},
		{"docx", config.DocTypeDOCX, false},
		{"word", config.DocTypeDOCX, false},
		{"WORD", config.DocTypeDOCX, false},
		{"xlsx", config.DocTypeXLSX, false},
		{"spreadsheet", config.DocTypeXLSX, false},
		{"sheet", config.DocTypeXLSX, false},
		{"XLSX", config.DocTypeXLSX, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := handler.normalizeDocType(tc.input)

			if tc.hasError {
				if err == nil {
					t.Errorf("Expected error for input '%s', got none", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for input '%s', got: %v", tc.input, err)
				}
				if result != tc.expected {
					t.Errorf("Expected '%s', got '%s'", tc.expected, result)
				}
			}
		})
	}
}

func TestDocumentHandler_GetContentType(t *testing.T) {
	handler := &DocumentHandler{}

	testCases := []struct {
		docType  string
		expected string
	}{
		{config.DocTypePDF, "application/pdf"},
		{config.DocTypeDOCX, "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{config.DocTypeXLSX, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		{"unknown", "application/octet-stream"},
	}

	for _, tc := range testCases {
		t.Run(tc.docType, func(t *testing.T) {
			result := handler.getContentType(tc.docType)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}
