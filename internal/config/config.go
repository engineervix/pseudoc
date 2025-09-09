package config

import (
	"errors"
	"time"
)

// Supported document types
const (
	DocTypePDF  = "pdf"
	DocTypeDOCX = "docx"
	DocTypeXLSX = "xlsx"
)

// Config holds the core document generation configuration
// This is shared between CLI and server modes
type Config struct {
	// Document type to generate
	DocType string

	// Document-specific options
	Pages  int // For PDF/DOCX
	Sheets int // For XLSX

	// Random seed for reproducible generation
	Seed int64

	// Generation options
	Count int // Number of documents to generate

	// Base filename (without extension, can be empty for auto-generation)
	BaseFilename string
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		DocType:      DocTypePDF,
		Pages:        1,
		Sheets:       1,
		Count:        1,
		BaseFilename: "", // Auto-generated
	}
}

// Validate validates the core document generation configuration
func (c *Config) Validate() error {
	if c.Count < 1 {
		return errors.New("count must be at least 1")
	}
	if c.Pages < 1 {
		return errors.New("pages must be at least 1")
	}
	if c.Sheets < 1 {
		return errors.New("sheets must be at least 1")
	}

	// Validate document type
	switch c.DocType {
	case DocTypePDF, DocTypeDOCX, DocTypeXLSX:
		// Valid types
	default:
		return errors.New("invalid document type")
	}

	return nil
}

// GenerateBaseFilename creates a base filename if none is provided
func (c *Config) GenerateBaseFilename() string {
	if c.BaseFilename != "" {
		return c.BaseFilename
	}

	// Generate timestamp-based filename
	return "pseudoc_" + time.Now().Format("2006-01-02_15-04-05")
}

// GetFileExtension returns the appropriate file extension for the document type
func (c *Config) GetFileExtension() string {
	switch c.DocType {
	case DocTypePDF:
		return ".pdf"
	case DocTypeDOCX:
		return ".docx"
	case DocTypeXLSX:
		return ".xlsx"
	default:
		return ".pdf" // fallback
	}
}
