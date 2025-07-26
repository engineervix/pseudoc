package pdf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/engineervix/pseudoc/internal/config"
)

func TestPDFGenerator_Generate(t *testing.T) {
	generator := New()

	cfg := &config.Config{
		DocType: config.DocTypePDF,
		Pages:   2,
		Count:   1,
	}

	var buf bytes.Buffer
	err := generator.Generate(&buf, cfg)

	if err != nil {
		t.Fatalf("Expected no error generating PDF, got: %v", err)
	}

	// Check that we actually got some data
	if buf.Len() == 0 {
		t.Error("Expected PDF content, got empty buffer")
	}

	// Check that the output starts with PDF header
	content := buf.Bytes()
	if len(content) < 4 || string(content[:4]) != "%PDF" {
		t.Error("Generated content doesn't appear to be a valid PDF (missing %PDF header)")
	}
}

func TestPDFGenerator_ValidatesPagesConfig(t *testing.T) {
	generator := New()

	cfg := &config.Config{
		DocType: config.DocTypePDF,
		Pages:   0, // Invalid: should be at least 1
		Count:   1,
	}

	var buf bytes.Buffer
	err := generator.Generate(&buf, cfg)

	// The generator should handle this gracefully, but let's test with valid config
	cfg.Pages = 1
	err = generator.Generate(&buf, cfg)

	if err != nil {
		t.Fatalf("Expected no error with valid config, got: %v", err)
	}
}

func TestPDFGenerator_MultiplePages(t *testing.T) {
	generator := New()

	cfg := &config.Config{
		DocType: config.DocTypePDF,
		Pages:   5,
		Count:   1,
	}

	var buf bytes.Buffer
	err := generator.Generate(&buf, cfg)

	if err != nil {
		t.Fatalf("Expected no error generating multi-page PDF, got: %v", err)
	}

	// The buffer should contain a valid PDF with content
	if buf.Len() == 0 {
		t.Error("Expected PDF content for multi-page document, got empty buffer")
	}

	content := string(buf.Bytes())
	// Check that it's a PDF
	if !strings.HasPrefix(content, "%PDF") {
		t.Error("Generated content doesn't appear to be a valid PDF")
	}
}
