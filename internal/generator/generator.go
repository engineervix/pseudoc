package generator

import (
	"fmt"
	"io"

	"github.com/engineervix/pseudoc/internal/config"
	"github.com/engineervix/pseudoc/internal/generator/pdf"
	"github.com/engineervix/pseudoc/internal/generator/xlsx"
)

// Generator defines the interface for document generators
type Generator interface {
	// Generate creates a document and writes it to the provided writer
	Generate(w io.Writer, cfg *config.Config) error
}

// GetGenerator returns the appropriate generator for the document type
func GetGenerator(docType string) (Generator, error) {
	switch docType {
	case config.DocTypePDF:
		return pdf.New(), nil
	case config.DocTypeDOCX:
		return &MockGenerator{DocType: "DOCX"}, nil
	case config.DocTypeXLSX:
		return xlsx.New(), nil
	default:
		return nil, fmt.Errorf("unsupported document type: %s", docType)
	}
}

// MockGenerator is a placeholder implementation that doesn't actually generate real documents
// It will be replaced with real implementations as we develop them
type MockGenerator struct {
	DocType string
}

// Generate implements the Generator interface but just writes a placeholder message
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
func getPageCount(docType string, cfg *config.Config) int {
	if docType == "XLSX" {
		return cfg.Sheets
	}
	return cfg.Pages
}
