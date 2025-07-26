package generator

import (
	"fmt"
	"io"

	"github.com/engineervix/pseudoc/internal/config"
	"github.com/engineervix/pseudoc/internal/generator/docx"
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
		return docx.New(), nil
	case config.DocTypeXLSX:
		return xlsx.New(), nil
	default:
		return nil, fmt.Errorf("unsupported document type: %s", docType)
	}
}
