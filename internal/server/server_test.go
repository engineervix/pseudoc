package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/engineervix/pseudoc/internal/config"
	"github.com/engineervix/pseudoc/internal/server/handlers"
)

func TestServer_Integration_GenerateEndpoints(t *testing.T) {
	// Setup server with test configuration
	cfg := config.DefaultServerConfig()
	cfg.EnableLogging = false // Disable logging for cleaner test output
	cfg.RateLimit = 0         // Disable rate limiting for tests

	server := New(cfg)
	server.setupMiddleware()
	server.setupRoutes()

	testCases := []struct {
		name                string
		method              string
		path                string
		body                string
		expectedStatus      int
		expectedContentType string
		checkBody           bool
	}{
		{
			name:                "GET PDF generation",
			method:              http.MethodGet,
			path:                "/api/v1/generate/pdf",
			expectedStatus:      http.StatusOK,
			expectedContentType: "application/pdf",
			checkBody:           true,
		},
		{
			name:                "GET DOCX generation",
			method:              http.MethodGet,
			path:                "/api/v1/generate/docx",
			expectedStatus:      http.StatusOK,
			expectedContentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			checkBody:           true,
		},
		{
			name:                "GET XLSX generation",
			method:              http.MethodGet,
			path:                "/api/v1/generate/xlsx",
			expectedStatus:      http.StatusOK,
			expectedContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			checkBody:           true,
		},
		{
			name:                "GET with query parameters",
			method:              http.MethodGet,
			path:                "/api/v1/generate/pdf?pages=2&seed=42",
			expectedStatus:      http.StatusOK,
			expectedContentType: "application/pdf",
			checkBody:           true,
		},
		{
			name:                "GET random document",
			method:              http.MethodGet,
			path:                "/api/v1/generate/random?seed=123",
			expectedStatus:      http.StatusOK,
			expectedContentType: "", // Will vary based on random selection
			checkBody:           true,
		},
		{
			name:           "GET invalid document type",
			method:         http.MethodGet,
			path:           "/api/v1/generate/invalid",
			expectedStatus: http.StatusBadRequest,
			checkBody:      false,
		},
		{
			name:                "POST PDF generation",
			method:              http.MethodPost,
			path:                "/api/v1/generate",
			body:                `{"type":"pdf","pages":2,"seed":42}`,
			expectedStatus:      http.StatusOK,
			expectedContentType: "application/pdf",
			checkBody:           true,
		},
		{
			name:                "POST random generation",
			method:              http.MethodPost,
			path:                "/api/v1/generate",
			body:                `{"type":"random","seed":456}`,
			expectedStatus:      http.StatusOK,
			expectedContentType: "", // Will vary
			checkBody:           true,
		},
		{
			name:           "POST invalid JSON",
			method:         http.MethodPost,
			path:           "/api/v1/generate",
			body:           `{"invalid":json}`,
			expectedStatus: http.StatusBadRequest,
			checkBody:      false,
		},
		{
			name:           "POST missing type",
			method:         http.MethodPost,
			path:           "/api/v1/generate",
			body:           `{"pages":2}`,
			expectedStatus: http.StatusBadRequest,
			checkBody:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}

			rec := httptest.NewRecorder()
			server.echo.ServeHTTP(rec, req)

			// Check status code
			if rec.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, rec.Code)
				if rec.Code >= 400 {
					t.Logf("Error response body: %s", rec.Body.String())
				}
			}

			// Check content type for successful requests
			if tc.expectedStatus == http.StatusOK && tc.expectedContentType != "" {
				contentType := rec.Header().Get("Content-Type")
				if contentType != tc.expectedContentType {
					t.Errorf("Expected Content-Type %s, got %s", tc.expectedContentType, contentType)
				}
			}

			// Check body content for successful requests
			if tc.checkBody && tc.expectedStatus == http.StatusOK {
				body := rec.Body.Bytes()
				if len(body) == 0 {
					t.Errorf("Expected document content, got empty body")
				}

				// Basic validation based on content type
				contentType := rec.Header().Get("Content-Type")
				switch contentType {
				case "application/pdf":
					if len(body) < 4 || string(body[:4]) != "%PDF" {
						t.Errorf("Expected PDF content, got invalid data")
					}
				case "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
					"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
					// DOCX and XLSX are ZIP files
					if len(body) < 2 || body[0] != 0x50 || body[1] != 0x4B {
						t.Errorf("Expected ZIP-based content (DOCX/XLSX), got invalid data")
					}
				}
			}
		})
	}
}

func TestServer_Integration_HealthAndInfo(t *testing.T) {
	cfg := config.DefaultServerConfig()
	cfg.EnableLogging = false
	cfg.RateLimit = 0 // Disable rate limiting for tests

	server := New(cfg)
	server.setupMiddleware()
	server.setupRoutes()

	// Test health endpoint
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	server.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected health check status 200, got %d", rec.Code)
	}

	var healthResp handlers.HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &healthResp); err != nil {
		t.Fatalf("Failed to parse health response: %v", err)
	}

	if healthResp.Status != "ok" {
		t.Errorf("Expected health status 'ok', got %s", healthResp.Status)
	}

	if healthResp.Version == "" {
		t.Errorf("Expected version to be set, got empty string")
	}

	if healthResp.Uptime == "" {
		t.Errorf("Expected uptime to be set, got empty string")
	}

	if healthResp.System == nil {
		t.Fatalf("Expected system info to be set, got nil")
	}

	if healthResp.System.GoroutineCount == 0 {
		t.Errorf("Expected goroutine count to be > 0, got 0")
	}

	// Test info endpoint
	req = httptest.NewRequest(http.MethodGet, "/api/v1/info", nil)
	rec = httptest.NewRecorder()
	server.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected info endpoint status 200, got %d", rec.Code)
	}

	var infoResp handlers.InfoResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &infoResp); err != nil {
		t.Fatalf("Failed to parse info response: %v", err)
	}

	if infoResp.Name != "pseudoc" {
		t.Errorf("Expected name 'pseudoc', got %s", infoResp.Name)
	}

	expectedFormats := []string{"pdf", "docx", "xlsx", "random"}
	if len(infoResp.SupportedFormats) != len(expectedFormats) {
		t.Errorf("Expected %d supported formats, got %d", len(expectedFormats), len(infoResp.SupportedFormats))
	}
}

func TestServer_Integration_CORS(t *testing.T) {
	cfg := config.DefaultServerConfig()
	cfg.EnableLogging = false
	cfg.CORSAllowedOrigins = []string{"http://example.com"}
	cfg.RateLimit = 0 // Disable rate limiting for tests

	server := New(cfg)
	server.setupMiddleware()
	server.setupRoutes()

	// Test CORS preflight request
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/generate/pdf", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")

	rec := httptest.NewRecorder()
	server.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected CORS preflight status 204, got %d", rec.Code)
	}

	// Check CORS headers
	corsHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "http://example.com",
		"Access-Control-Allow-Methods": "",
		"Access-Control-Allow-Headers": "",
	}

	for header, expectedValue := range corsHeaders {
		value := rec.Header().Get(header)
		if value == "" {
			t.Errorf("Expected CORS header %s, got empty", header)
		}
		if expectedValue != "" && value != expectedValue {
			t.Errorf("Expected CORS header %s to be %s, got %s", header, expectedValue, value)
		}
	}
}

func TestServer_Integration_RateLimit(t *testing.T) {
	// Test 1: Verify that rate limiting can be disabled
	cfgNoLimit := config.DefaultServerConfig()
	cfgNoLimit.EnableLogging = false
	cfgNoLimit.RateLimit = 0 // Disabled

	serverNoLimit := New(cfgNoLimit)
	serverNoLimit.setupMiddleware()
	serverNoLimit.setupRoutes()

	// Should handle many requests without rate limiting
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		serverNoLimit.echo.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected request %d without rate limiting to succeed, got %d", i+1, rec.Code)
		}
	}

	// Test 2: Verify that rate limiting can be enabled (configuration test)
	cfgWithLimit := config.DefaultServerConfig()
	cfgWithLimit.EnableLogging = false
	cfgWithLimit.RateLimit = 60 // Normal limit

	serverWithLimit := New(cfgWithLimit)
	serverWithLimit.setupMiddleware()
	serverWithLimit.setupRoutes()

	// Make one request to verify the server works
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	serverWithLimit.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected server with rate limiting to handle single request, got %d", rec.Code)
	}
}

func TestServer_Integration_ErrorHandling(t *testing.T) {
	cfg := config.DefaultServerConfig()
	cfg.EnableLogging = false
	cfg.RateLimit = 0 // Disable rate limiting for tests

	server := New(cfg)
	server.setupMiddleware()
	server.setupRoutes()

	errorTestCases := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
		checkErrorJSON bool
	}{
		{
			name:           "Invalid document type",
			method:         http.MethodGet,
			path:           "/api/v1/generate/invalid",
			expectedStatus: http.StatusBadRequest,
			checkErrorJSON: true,
		},
		{
			name:           "Invalid query parameters",
			method:         http.MethodGet,
			path:           "/api/v1/generate/pdf?pages=invalid",
			expectedStatus: http.StatusBadRequest,
			checkErrorJSON: true,
		},
		{
			name:           "Missing required field in POST",
			method:         http.MethodPost,
			path:           "/api/v1/generate",
			body:           `{"pages":2}`,
			expectedStatus: http.StatusBadRequest,
			checkErrorJSON: true,
		},
		{
			name:           "Not found endpoint",
			method:         http.MethodGet,
			path:           "/api/v1/nonexistent",
			expectedStatus: http.StatusNotFound,
			checkErrorJSON: false,
		},
	}

	for _, tc := range errorTestCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}

			rec := httptest.NewRecorder()
			server.echo.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, rec.Code)
			}

			if tc.checkErrorJSON {
				var errorResp handlers.ErrorResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &errorResp); err != nil {
					t.Errorf("Expected JSON error response, got: %v", err)
				} else {
					if errorResp.Error == "" {
						t.Errorf("Expected error message in response")
					}
					if errorResp.Code != tc.expectedStatus {
						t.Errorf("Expected error code %d in response, got %d", tc.expectedStatus, errorResp.Code)
					}
				}
			}
		})
	}
}
