package docx

import (
	"bytes"
	"testing"

	"github.com/engineervix/pseudoc/internal/config"
)

func TestDOCXGenerator_Generate(t *testing.T) {
	generator := New()

	cfg := &config.Config{
		DocType: config.DocTypeDOCX,
		Pages:   2,
		Count:   1,
	}

	var buf bytes.Buffer
	err := generator.Generate(&buf, cfg)

	if err != nil {
		t.Fatalf("Expected no error generating DOCX, got: %v", err)
	}

	// Check that we actually got some data
	if buf.Len() == 0 {
		t.Error("Expected DOCX content, got empty buffer")
	}

	// Check that the output contains some DOCX-like content
	// DOCX files are ZIP archives, so they should start with the ZIP signature
	content := buf.Bytes()
	if len(content) < 4 {
		t.Error("Generated content is too short to be a valid DOCX file")
	}

	// ZIP files start with "PK" (0x50, 0x4B)
	if len(content) >= 2 && (content[0] != 0x50 || content[1] != 0x4B) {
		t.Error("Generated content doesn't appear to be a valid DOCX file (missing ZIP signature)")
	}
}

func TestDOCXGenerator_ValidatesPagesConfig(t *testing.T) {
	generator := New()

	cfg := &config.Config{
		DocType: config.DocTypeDOCX,
		Pages:   1, // Valid minimum
		Count:   1,
	}

	var buf bytes.Buffer
	err := generator.Generate(&buf, cfg)

	if err != nil {
		t.Fatalf("Expected no error with valid config, got: %v", err)
	}

	// Verify we got content
	if buf.Len() == 0 {
		t.Error("Expected DOCX content with valid config, got empty buffer")
	}
}

func TestDOCXGenerator_MultiplePages(t *testing.T) {
	generator := New()

	cfg := &config.Config{
		DocType: config.DocTypeDOCX,
		Pages:   5,
		Count:   1,
	}

	var buf bytes.Buffer
	err := generator.Generate(&buf, cfg)

	if err != nil {
		t.Fatalf("Expected no error generating multi-page DOCX, got: %v", err)
	}

	// The buffer should contain a valid DOCX with content
	if buf.Len() == 0 {
		t.Error("Expected DOCX content for multi-page document, got empty buffer")
	}

	content := buf.Bytes()
	// Check ZIP signature
	if len(content) >= 2 && (content[0] != 0x50 || content[1] != 0x4B) {
		t.Error("Generated multi-page content doesn't appear to be a valid DOCX file")
	}
}

func TestSplitIntoWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"simple text", "hello world", []string{"hello", "world"}},
		{"multiple spaces", "hello   world", []string{"hello", "world"}},
		{"single word", "hello", []string{"hello"}},
		{"empty string", "", []string{}},
		{"spaces only", "   ", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitIntoWords(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d words, got %d", len(tt.expected), len(result))
				return
			}

			for i, word := range result {
				if word != tt.expected[i] {
					t.Errorf("Expected word %d to be '%s', got '%s'", i, tt.expected[i], word)
				}
			}
		})
	}
}

func TestJoinWords(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{"two words", []string{"hello", "world"}, "hello world"},
		{"single word", []string{"hello"}, "hello"},
		{"empty slice", []string{}, ""},
		{"three words", []string{"one", "two", "three"}, "one two three"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinWords(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
