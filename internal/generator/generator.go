package generator

import (
	"fmt"
	"io"

	"github.com/engineervix/pseudoc/internal/config"
)

// Generator defines the interface for generating documents
type Generator interface {
	// Generate creates a document and writes it to the provided writer
	Generate(w io.Writer, cfg *config.Config) error
}

// GetGenerator returns the appropriate generator for the document type
func GetGenerator(docType string) (Generator, error) {
	switch docType {
	// NOTE: We are returning a pointer instead of a value because
	// 1. The Generate method is implemented with a pointer receiver
	// 2.  If a method has a pointer receiver (g *MockGenerator), then
	//     only pointer types *MockGenerator satisfy the interface, not the
	//     value type MockGenerator.
	case config.DocTypePDF:
		return &MockGenerator{DocType: "PDF"}, nil
	case config.DocTypeDOCX:
		return &MockGenerator{DocType: "DOCX"}, nil
	case config.DocTypeXLSX:
		return &MockGenerator{DocType: "XLSX"}, nil
	default:
		// NOTE: In Go, nil is a valid zero value for any interface type
		// fmt.Errorf() creates and returns an error type
		return nil, fmt.Errorf("Unsurpotted document type: %s", docType)
	}
}

// MockGenerator is a placeholder implementation that doesn't actually generate real documents.
// I'll replace it with real implementations as I develop them
type MockGenerator struct {
	DocType string
}

// Generate implements the Generator interface, but for now we just write a placeholder msg
func (g *MockGenerator) Generate(w io.Writer, cfg *config.Config) error {
	msg := fmt.Sprintf(
		"This is a mock %s document\nPages/Sheets: %d\n",
		g.DocType,
		getPageCount(g.DocType, cfg),
	)

	_, err := w.Write([]byte(msg))
	return err
}

// getPageCount returns the appropriate page or sheet count based on document type
func getPageCount(docType string, c *config.Config) int {
	if docType == "XLSX" {
		return c.Sheets
	}

	return c.Pages
}
