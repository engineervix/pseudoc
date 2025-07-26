package xlsx

import (
	"bytes"
	"testing"

	"github.com/engineervix/pseudoc/internal/config"
)

func TestXLSXGenerator_Generate(t *testing.T) {
	generator := New()

	cfg := &config.Config{
		DocType: config.DocTypeXLSX,
		Sheets:  2,
		Count:   1,
	}

	var buf bytes.Buffer
	err := generator.Generate(&buf, cfg)

	if err != nil {
		t.Fatalf("Expected no error generating XLSX, got: %v", err)
	}

	// Check that we actually got some data
	if buf.Len() == 0 {
		t.Error("Expected XLSX content, got empty buffer")
	}

	// Check that the output contains some XLSX-like content
	// XLSX files are ZIP archives, so they should start with the ZIP signature
	content := buf.Bytes()
	if len(content) < 4 {
		t.Error("Generated content is too short to be a valid XLSX file")
	}

	// ZIP files start with "PK" (0x50, 0x4B)
	if len(content) >= 2 && (content[0] != 0x50 || content[1] != 0x4B) {
		t.Error("Generated content doesn't appear to be a valid XLSX file (missing ZIP signature)")
	}
}

func TestXLSXGenerator_ValidatesSheetsConfig(t *testing.T) {
	generator := New()

	cfg := &config.Config{
		DocType: config.DocTypeXLSX,
		Sheets:  1, // Valid minimum
		Count:   1,
	}

	var buf bytes.Buffer
	err := generator.Generate(&buf, cfg)

	if err != nil {
		t.Fatalf("Expected no error with valid config, got: %v", err)
	}

	// Verify we got content
	if buf.Len() == 0 {
		t.Error("Expected XLSX content with valid config, got empty buffer")
	}
}

func TestXLSXGenerator_MultipleSheets(t *testing.T) {
	generator := New()

	cfg := &config.Config{
		DocType: config.DocTypeXLSX,
		Sheets:  5,
		Count:   1,
	}

	var buf bytes.Buffer
	err := generator.Generate(&buf, cfg)

	if err != nil {
		t.Fatalf("Expected no error generating multi-sheet XLSX, got: %v", err)
	}

	// The buffer should contain a valid XLSX with content
	if buf.Len() == 0 {
		t.Error("Expected XLSX content for multi-sheet document, got empty buffer")
	}

	content := buf.Bytes()
	// Check ZIP signature
	if len(content) >= 2 && (content[0] != 0x50 || content[1] != 0x4B) {
		t.Error("Generated multi-sheet content doesn't appear to be a valid XLSX file")
	}
}
