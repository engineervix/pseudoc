package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/engineervix/pseudoc/internal/config"
	"github.com/engineervix/pseudoc/internal/server/handlers"
	"github.com/labstack/echo/v4"
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
			// Special handling for the random GET endpoint to test redirects
			if tc.name == "GET random document" {
				// Create a test server
				testServer := httptest.NewServer(server.echo)
				defer testServer.Close()

				// Create a client that follows redirects
				client := &http.Client{
					CheckRedirect: func(req *http.Request, via []*http.Request) error {
						// Allow redirects
						return nil
					},
				}

				// Make the request to the test server
				res, err := client.Get(testServer.URL + tc.path)
				if err != nil {
					t.Fatalf("Failed to make request to test server: %v", err)
				}
				defer res.Body.Close()

				// The final response after redirect should be 200 OK
				if res.StatusCode != http.StatusOK {
					t.Errorf("Expected status 200 after redirect, got %d", res.StatusCode)
				}

				// Check that we received a valid document
				if res.ContentLength == 0 {
					t.Errorf("Expected document content, got empty body")
				}
				return // End this test case here
			}
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

	// Verify that limits use config constants, not hardcoded values
	if infoResp.Limits.MaxPages != config.MaxPages {
		t.Errorf("Expected MaxPages to be %d (from config.MaxPages), got %d", config.MaxPages, infoResp.Limits.MaxPages)
	}
	if infoResp.Limits.MaxSheets != config.MaxSheets {
		t.Errorf("Expected MaxSheets to be %d (from config.MaxSheets), got %d", config.MaxSheets, infoResp.Limits.MaxSheets)
	}
}

func TestServer_Integration_CORS_Disabled_By_Default(t *testing.T) {
	cfg := config.DefaultServerConfig()
	cfg.EnableLogging = false
	cfg.RateLimit = 0 // Disable rate limiting for tests
	// Note: CORSAllowedOrigins should be empty by default

	server := New(cfg)
	server.setupMiddleware()
	server.setupRoutes()

	// Test that CORS headers are not present when CORS is disabled
	req := httptest.NewRequest(http.MethodGet, "/api/v1/generate/pdf", nil)
	req.Header.Set("Origin", "http://example.com")

	rec := httptest.NewRecorder()
	server.echo.ServeHTTP(rec, req)

	// Should not have CORS headers when CORS is disabled
	corsHeaders := []string{
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers",
	}

	for _, header := range corsHeaders {
		if value := rec.Header().Get(header); value != "" {
			t.Errorf("Expected no CORS header %s when CORS is disabled, got %s", header, value)
		}
	}
}

func TestServer_Integration_CORS_Enabled(t *testing.T) {
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

func TestServer_Integration_RateLimit_Applied_Correctly(t *testing.T) {
	cfg := config.DefaultServerConfig()
	cfg.EnableLogging = false
	cfg.RateLimit = 1 // Very low rate limit for testing

	server := New(cfg)
	server.setupMiddleware()
	server.setupRoutes()

	t.Run("RateLimit_DocumentGeneration", func(t *testing.T) {
		// Test that actual document generation endpoints are rate limited
		endpoint := "/api/v1/generate/pdf"

		// Make multiple requests quickly to trigger rate limiting
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodGet, endpoint, nil)
			rec := httptest.NewRecorder()
			server.echo.ServeHTTP(rec, req)

			// After the first few requests, we should get rate limited
			if i > 2 && rec.Code == http.StatusTooManyRequests {
				// Verify the error response format
				var errorResp map[string]string
				if err := json.Unmarshal(rec.Body.Bytes(), &errorResp); err != nil {
					t.Errorf("Failed to parse rate limit error response: %v", err)
				}
				if errorResp["error"] != "rate limit exceeded" {
					t.Errorf("Expected rate limit error, got %s", errorResp["error"])
				}
				return // Test passed - rate limiting is working
			}
		}
		t.Errorf("Expected rate limiting to be triggered for endpoint %s", endpoint)
	})

	t.Run("RateLimit_RandomRedirect_NotLimited", func(t *testing.T) {
		// Test that random endpoint redirects are not rate limited
		endpoint := "/api/v1/generate/random"

		// Make multiple requests - should all get redirects, not rate limited
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodGet, endpoint+"?seed="+fmt.Sprintf("%d", i), nil)
			rec := httptest.NewRecorder()
			server.echo.ServeHTTP(rec, req)

			// Should get redirects, not rate limit errors
			if rec.Code == http.StatusTooManyRequests {
				t.Errorf("Random endpoint should not be rate limited on redirect, got rate limit error on request %d", i)
				return
			}
			if rec.Code != http.StatusTemporaryRedirect {
				t.Errorf("Expected redirect (307) for random endpoint, got %d", rec.Code)
				return
			}
		}
	})
}

func TestServer_Integration_IPExtractor(t *testing.T) {
	cfg := config.DefaultServerConfig()
	cfg.EnableLogging = false
	cfg.RateLimit = 0 // Disable rate limiting for this test

	server := New(cfg)
	server.setupMiddleware()

	// Add a test endpoint that returns the detected IP
	server.echo.GET("/test-ip", func(c echo.Context) error {
		return c.String(http.StatusOK, c.RealIP())
	})

	testCases := []struct {
		name       string
		headers    map[string]string
		expectedIP string
	}{
		{
			name: "Cloudflare IP",
			headers: map[string]string{
				"CF-Connecting-IP": "198.51.100.1",
				"X-Real-IP":        "203.0.113.1",
				"X-Forwarded-For":  "192.0.2.1",
			},
			expectedIP: "198.51.100.1",
		},
		{
			name: "X-Real-IP",
			headers: map[string]string{
				"X-Real-IP":       "203.0.113.1",
				"X-Forwarded-For": "192.0.2.1",
			},
			expectedIP: "203.0.113.1",
		},
		{
			name: "X-Forwarded-For",
			headers: map[string]string{
				"X-Forwarded-For": "192.0.2.1, 203.0.113.1",
			},
			expectedIP: "192.0.2.1",
		},
		{
			name:       "Direct IP",
			headers:    map[string]string{},
			expectedIP: "192.0.2.1", // Default from httptest
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test-ip", nil)
			for key, value := range tc.headers {
				req.Header.Set(key, value)
			}
			rec := httptest.NewRecorder()
			server.echo.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rec.Code)
			}

			// httptest sets RemoteAddr to 192.0.2.1:1234 by default
			body := rec.Body.String()
			expected := tc.expectedIP

			if tc.name == "Direct IP" {
				ip, _, err := net.SplitHostPort(body)
				if err != nil {
					// If no port, body is just the IP
					ip = body
				}
				if ip != expected {
					t.Errorf("Expected IP '%s', got '%s'", expected, ip)
				}
			} else {
				if body != expected {
					t.Errorf("Expected IP '%s', got '%s'", expected, body)
				}
			}
		})
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
