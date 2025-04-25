package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// Supported document types
const (
	DocTypePDF  = "pdf"
	DocTypeDOCX = "docx"
	DocTypeXLSX = "xlsx"
)

// Config holds the application configuration
type Config struct {
	// Document type to generate
	DocType string

	//  Output options
	OutputDir string
	Filename  string
	Count     int

	// Document-specific options
	Pages  int
	Sheets int
	Size   string

	// Random seed for reproducible generation
	Seed int64
}

func DefaultConfig() Config {
	return Config{
		DocType:   DocTypePDF,
		OutputDir: ".",
		Count:     1,
		Pages:     1,
		Sheets:    1,
	}
}

func (c *Config) Validate() error {
	if c.Count < 1 {
		return errors.New("Count must be at least 1")
	}
	if c.Pages < 1 {
		return errors.New("Pages must be at least 1")
	}
	if c.Sheets < 1 {
		return errors.New("Sheets must be at least 1")
	}

	if err := validateOutputDir(c.OutputDir); err != nil {
		return err
	}

	return nil
}

// Checks if the output directory exists and creates it if needed
func validateOutputDir(dir string) error {
	// check if directory exists
	info, err := os.Stat(dir)
	if err == nil {
		// path exists, check if it's a directory
		if !info.IsDir() {
			return errors.New("output path exists but isn't a directory")
		}
		return nil
	}

	// If the error isn't that the directory doesn't exist, return the error

	// NOTE: New code should use this form ...
	if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	// ... the alternative (older, pre Go 1.13) way of doing it is
	// if !os.IsNotExist(err) {
	// 	return err
	// }

	// Create the directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return nil
}

// Creates a filename for the document
func (c *Config) GenerateFileName(index int) string {
	base := c.Filename
	if base == "" {
		// NOTE: Go's date formatting uses a reference date (January 2, 2006 at 15:04:05 MST)
		// as a template rather than format specifiers.
		// The reference date was specifically chosen to be 01/02 03:04:05 PM '06 -0700
		// (or in other formats, 1/2/2006 at 3:04:05PM in MST timezone).
		// This was a cute way to use the sequence 1, 2, 3, 4, 5, 6, 7 to represent different date and time components:
		// month(1), day(2), hour(3), minute(4), second(5), year(6), timezone(7).
		// It's sometimes called the "magic date" among Go programmers because it seems odd at first,
		// but becomes quite intuitive once you get used to it.
		base = "pseudoc_" + time.Now().Format("2006-01-02_15-04-05")
	}

	// Add index suffix if adding multiple documents
	// NOTE: in Go, character literals like 'a' are actually numeric values of type rune
	// (which is an alias for int32), not a separate "char" type like in some other languages.
	if c.Count > 1 {
		base += "_" + string(rune('a'+index))
	}

	// Add extension based on document type
	var ext string
	switch c.DocType {
	case DocTypePDF:
		ext = ".pdf"
	case DocTypeDOCX:
		ext = ".docx"
	case DocTypeXLSX:
		ext = ".xlsx"
	default:
		// use PDF as default
		ext = ".pdf"
	}

	return filepath.Join(c.OutputDir, base+ext)

}
