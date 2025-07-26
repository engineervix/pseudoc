package docx

import (
	"fmt"
	"io"
	"math/rand"

	"github.com/gomutex/godocx"
	godocxdoc "github.com/gomutex/godocx/docx"

	appconfig "github.com/engineervix/pseudoc/internal/config"
	"github.com/engineervix/pseudoc/internal/content"
)

// Generator implements the DOCX document generator using gomutex/godocx library
type Generator struct{}

// New creates a new DOCX generator
func New() *Generator {
	return &Generator{}
}

// Generate creates a DOCX document and writes it to the provided writer
func (g *Generator) Generate(w io.Writer, cfg *appconfig.Config) error {
	// Create a new document
	document, err := godocx.NewDocument()
	if err != nil {
		return fmt.Errorf("failed to create new document: %w", err)
	}

	// Generate the requested number of pages
	for pageNum := 1; pageNum <= cfg.Pages; pageNum++ {
		if pageNum > 1 {
			// Add a page break before each new page (except the first)
			document.AddPageBreak()
		}

		g.addPageContent(document, pageNum)
	}

	// Write the document to the provided writer
	return document.Write(w)
}

// addPageContent adds content to simulate a page
func (g *Generator) addPageContent(document *godocxdoc.RootDoc, pageNum int) {
	// Add page title as a heading
	pageTitle := fmt.Sprintf("Page %d - %s", pageNum, content.GenerateTitle())
	document.AddHeading(pageTitle, 1)

	// Add introduction paragraph
	introText := content.GenerateParagraph(rand.Intn(80) + 40) // 40-120 words
	document.AddParagraph(introText)

	// Add a subheading
	subheading := content.GenerateTitle()
	document.AddHeading(subheading, 2)

	// Add content paragraphs
	numParagraphs := rand.Intn(3) + 2 // 2-4 paragraphs per page
	for i := 0; i < numParagraphs; i++ {
		paraText := content.GenerateParagraph(rand.Intn(100) + 50) // 50-150 words

		// Add paragraph with mixed formatting
		p := document.AddParagraph("")
		words := splitIntoWords(paraText)
		wordCount := len(words)

		if wordCount > 0 {
			// Add first part as regular text
			firstPart := joinWords(words[:wordCount/3])
			p.AddText(firstPart)

			// Add middle part with bold formatting
			if wordCount > 3 {
				middlePart := joinWords(words[wordCount/3 : 2*wordCount/3])
				r1 := p.AddText(" " + middlePart)
				r1.Bold(true)
			}

			// Add last part with italic formatting
			if wordCount > 6 {
				lastPart := joinWords(words[2*wordCount/3:])
				r2 := p.AddText(" " + lastPart)
				r2.Italic(true)
			}
		}
	}

	// Add a bullet list sometimes
	if rand.Float32() < 0.6 { // 60% chance of adding a list
		titlePara := document.AddParagraph("Key Points:")
		titleRun := titlePara.AddText("")
		titleRun.Bold(true)

		listItems := content.GenerateListItems(rand.Intn(4) + 3) // 3-6 items
		for _, item := range listItems {
			listPara := document.AddParagraph(item)
			listPara.Style("List Bullet")
		}
	}

	// Add a simple table sometimes
	if rand.Float32() < 0.4 { // 40% chance of adding a table
		g.addSimpleTable(document)
	}
}

// addSimpleTable adds a simple table to the document
func (g *Generator) addSimpleTable(document *godocxdoc.RootDoc) {
	// Add table title
	titlePara := document.AddParagraph("Data Summary")
	titleRun := titlePara.AddText("")
	titleRun.Bold(true)

	// Generate table data
	rows := rand.Intn(4) + 3 // 3-6 rows
	cols := rand.Intn(2) + 3 // 3-4 columns
	tableData := content.GenerateTableData(rows, cols)

	// Create table
	table := document.AddTable()
	table.Style("LightList-Accent3")

	// Add header row
	headers := []string{"Item", "Value", "Status", "Notes"}
	headerRow := table.AddRow()
	for i := 0; i < cols && i < len(headers); i++ {
		cell := headerRow.AddCell()
		headerPara := cell.AddParagraph(headers[i])
		headerRun := headerPara.AddText("")
		headerRun.Bold(true)
	}

	// Add data rows
	for _, rowData := range tableData {
		dataRow := table.AddRow()
		for i := 0; i < cols && i < len(rowData); i++ {
			cell := dataRow.AddCell()
			cell.AddParagraph(rowData[i])
		}
	}
}

// Helper functions for text manipulation
func splitIntoWords(text string) []string {
	words := []string{}
	current := ""

	for _, char := range text {
		if char == ' ' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		words = append(words, current)
	}

	return words
}

func joinWords(words []string) string {
	if len(words) == 0 {
		return ""
	}

	result := words[0]
	for i := 1; i < len(words); i++ {
		result += " " + words[i]
	}

	return result
}
