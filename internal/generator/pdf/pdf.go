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

	// Generate content based on the number of pages requested
	for pageNum := 1; pageNum <= cfg.Pages; pageNum++ {
		// Add a new page for each page after the first
		if pageNum > 1 {
			// Create and add a new page
			m.AddPages(page.New())
		}

		if err := g.addPage(m, pageNum, cfg); err != nil {
			return fmt.Errorf("failed to add page %d: %w", pageNum, err)
		}
	}

	// Generate the PDF and write to the provided writer
	document, err := m.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate PDF: %w", err)
	}

	_, err = w.Write(document.GetBytes())
	return err
}

// addPage adds a single page with placeholder content to the PDF
func (g *Generator) addPage(m core.Maroto, pageNum int, cfg *appconfig.Config) error {
	// Add page header
	g.addHeader(m, pageNum)

	// Add some paragraphs of Lorem Ipsum text
	paragraphCount := rand.Intn(4) + 2 // 2-5 paragraphs per page
	for i := 0; i < paragraphCount; i++ {
		g.addParagraph(m)
	}

	// Occasionally add a table or list
	if rand.Float32() < 0.3 { // 30% chance
		if rand.Float32() < 0.5 {
			g.addTable(m)
		} else {
			g.addList(m)
		}
	}

	// Add some spacing at the end
	m.AddRow(10)

	return nil
}

// addHeader adds a header section to the page
func (g *Generator) addHeader(m core.Maroto, pageNum int) {
	// Main title (only on first page)
	if pageNum == 1 {
		m.AddRow(20,
			col.New(12).Add(
				text.New("Pseudoc Generated Document", props.Text{
					Top:   5,
					Style: fontstyle.Bold,
					Align: align.Center,
					Size:  16,
				}),
			),
		)

		m.AddRow(5) // spacing
	}

	// Page subtitle
	subtitle := fmt.Sprintf("Page %d - %s", pageNum, content.GenerateTitle())
	m.AddRow(15,
		col.New(12).Add(
			text.New(subtitle, props.Text{
				Top:   3,
				Style: fontstyle.Bold,
				Size:  12,
			}),
		),
	)

	m.AddRow(5) // spacing after header
}

// addParagraph adds a paragraph of Lorem Ipsum text
func (g *Generator) addParagraph(m core.Maroto) {
	paragraphText := content.GenerateParagraph(rand.Intn(100) + 50) // 50-150 words

	m.AddRow(25,
		col.New(12).Add(
			text.New(paragraphText, props.Text{
				Top:   3,
				Align: align.Left,
				Size:  10,
			}),
		),
	)

	m.AddRow(5) // spacing between paragraphs
}

// addTable adds a simple table with placeholder data
func (g *Generator) addTable(m core.Maroto) {
	// Table header
	m.AddRow(15,
		col.New(3).Add(
			text.New("Item", props.Text{
				Top:   3,
				Style: fontstyle.Bold,
				Size:  10,
			}),
		),
		col.New(6).Add(
			text.New("Description", props.Text{
				Top:   3,
				Style: fontstyle.Bold,
				Size:  10,
			}),
		),
		col.New(3).Add(
			text.New("Value", props.Text{
				Top:   3,
				Style: fontstyle.Bold,
				Align: align.Right,
				Size:  10,
			}),
		),
	)

	// Table rows
	rowCount := rand.Intn(5) + 3 // 3-7 rows
	for i := 0; i < rowCount; i++ {
		item := content.GenerateWord()
		description := content.GenerateSentence(rand.Intn(8) + 3) // 3-10 words
		value := fmt.Sprintf("$%.2f", rand.Float64()*1000)

		m.AddRow(12,
			col.New(3).Add(
				text.New(item, props.Text{
					Top:  2,
					Size: 9,
				}),
			),
			col.New(6).Add(
				text.New(description, props.Text{
					Top:  2,
					Size: 9,
				}),
			),
			col.New(3).Add(
				text.New(value, props.Text{
					Top:   2,
					Align: align.Right,
					Size:  9,
				}),
			),
		)
	}

	m.AddRow(10) // spacing after table
}

// addList adds a bulleted list
func (g *Generator) addList(m core.Maroto) {
	// List title
	m.AddRow(15,
		col.New(12).Add(
			text.New("Key Points:", props.Text{
				Top:   3,
				Style: fontstyle.Bold,
				Size:  11,
			}),
		),
	)

	// List items
	itemCount := rand.Intn(4) + 3 // 3-6 items
	for i := 0; i < itemCount; i++ {
		listItem := fmt.Sprintf("• %s", content.GenerateSentence(rand.Intn(10)+5)) // 5-14 words

		m.AddRow(12,
			col.New(12).Add(
				text.New(listItem, props.Text{
					Top:  2,
					Size: 10,
				}),
			),
		)
	}

	m.AddRow(10) // spacing after list
}
