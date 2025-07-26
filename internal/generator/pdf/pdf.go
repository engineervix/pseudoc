package pdf

import (
	"fmt"
	"io"
	"math/rand"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/page"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"

	appconfig "github.com/engineervix/pseudoc/internal/config"
	"github.com/engineervix/pseudoc/internal/content"
)

// Generator implements the PDF document generator using maroto
type Generator struct{}

// New creates a new PDF generator
func New() *Generator {
	return &Generator{}
}

// Generate creates a PDF document and writes it to the provided writer
func (g *Generator) Generate(w io.Writer, cfg *appconfig.Config) error {
	// Create maroto configuration
	marotoCfg := config.NewBuilder().
		WithPageSize(pagesize.A4).
		WithOrientation(orientation.Vertical).
		Build()

	// Create maroto instance
	m := maroto.New(marotoCfg)

	// Generate exactly the requested number of pages
	for pageNum := 1; pageNum <= cfg.Pages; pageNum++ {
		// Add a new page for each page after the first
		if pageNum > 1 {
			m.AddPages(page.New())
		}

		g.addSimplePage(m, pageNum)
	}

	// Generate the PDF and write to the provided writer
	document, err := m.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate PDF: %w", err)
	}

	_, err = w.Write(document.GetBytes())
	return err
}

// Remove the old addPage method since we're not using page-based generation anymore

// addSimplePage adds minimal content to a page - just a heading and a paragraph
func (g *Generator) addSimplePage(m core.Maroto, pageNum int) {
	// Page title
	pageTitle := fmt.Sprintf("Page %d - %s", pageNum, content.GenerateTitle())
	m.AddRow(20,
		col.New(12).Add(
			text.New(pageTitle, props.Text{
				Top:   5,
				Style: fontstyle.Bold,
				Size:  14,
			}),
		),
	)

	// Spacing after title
	m.AddRow(10)

	// Single paragraph of Lorem Ipsum text
	paragraphText := content.GenerateParagraph(rand.Intn(100) + 50) // 50-150 words
	m.AddRow(30,
		col.New(12).Add(
			text.New(paragraphText, props.Text{
				Top:   5,
				Align: align.Left,
				Size:  11,
			}),
		),
	)
}
