package xlsx

import (
	"fmt"
	"io"
	"math/rand"

	"github.com/xuri/excelize/v2"

	appconfig "github.com/engineervix/pseudoc/internal/config"
	"github.com/engineervix/pseudoc/internal/content"
)

// Generator implements the XLSX document generator using excelize
type Generator struct{}

// New creates a new XLSX generator
func New() *Generator {
	return &Generator{}
}

// Generate creates an XLSX document and writes it to the provided writer
func (g *Generator) Generate(w io.Writer, cfg *appconfig.Config) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Warning: failed to close Excel file: %v\n", err)
		}
	}()

	// Generate the requested number of sheets
	for sheetNum := 1; sheetNum <= cfg.Sheets; sheetNum++ {
		sheetName := fmt.Sprintf("Sheet%d", sheetNum)

		// Create new sheet (first sheet already exists)
		if sheetNum > 1 {
			_, err := f.NewSheet(sheetName)
			if err != nil {
				return fmt.Errorf("failed to create sheet %s: %w", sheetName, err)
			}
		} else {
			// Rename the default sheet
			err := f.SetSheetName("Sheet1", sheetName)
			if err != nil {
				return fmt.Errorf("failed to rename first sheet: %w", err)
			}
		}

		// Add content to the sheet
		if err := g.populateSheet(f, sheetName, sheetNum); err != nil {
			return fmt.Errorf("failed to populate sheet %s: %w", sheetName, err)
		}
	}

	// Write to the provided writer
	return f.Write(w)
}

// populateSheet adds content to a specific sheet
func (g *Generator) populateSheet(f *excelize.File, sheetName string, sheetNum int) error {
	// Add a title
	title := fmt.Sprintf("Data Sheet %d - %s", sheetNum, content.GenerateTitle())
	if err := f.SetCellValue(sheetName, "A1", title); err != nil {
		return err
	}

	// Style the title
	titleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 16,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	if err != nil {
		return err
	}

	if err := f.SetCellStyle(sheetName, "A1", "E1", titleStyle); err != nil {
		return err
	}

	// Merge cells for title
	if err := f.MergeCell(sheetName, "A1", "E1"); err != nil {
		return err
	}

	// Add some spacing
	currentRow := 3

	// Generate a data table
	tableRows := rand.Intn(15) + 10 // 10-24 rows
	tableCols := rand.Intn(3) + 3   // 3-5 columns

	// Add table headers
	headers := []string{"Company", "Product", "Revenue", "Growth Rate", "Notes"}
	for col := 0; col < tableCols && col < len(headers); col++ {
		cell := fmt.Sprintf("%s%d", string(rune('A'+col)), currentRow)
		if err := f.SetCellValue(sheetName, cell, headers[col]); err != nil {
			return err
		}
	}

	// Style headers
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E8E8E8"},
			Pattern: 1,
		},
	})
	if err != nil {
		return err
	}

	// Apply header style to each cell individually to avoid range issues
	for col := 0; col < tableCols; col++ {
		cell := fmt.Sprintf("%s%d", string(rune('A'+col)), currentRow)
		if err := f.SetCellStyle(sheetName, cell, cell, headerStyle); err != nil {
			return err
		}
	}

	currentRow++

	// Generate table data
	tableData := content.GenerateTableData(tableRows, tableCols)
	for rowIdx, row := range tableData {
		for colIdx, cellValue := range row {
			if colIdx >= tableCols {
				break
			}
			cell := fmt.Sprintf("%s%d", string(rune('A'+colIdx)), currentRow+rowIdx)
			if err := f.SetCellValue(sheetName, cell, cellValue); err != nil {
				return err
			}
		}
	}

	// Auto-adjust column widths
	for col := 0; col < tableCols; col++ {
		colName := string(rune('A' + col))
		if err := f.SetColWidth(sheetName, colName, colName, 15); err != nil {
			return err
		}
	}

	// Add some additional data sections
	currentRow += len(tableData) + 3

	// Add a summary section
	summaryTitle := "Summary Statistics"
	summaryCell := fmt.Sprintf("A%d", currentRow)
	if err := f.SetCellValue(sheetName, summaryCell, summaryTitle); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheetName, summaryCell, summaryCell, titleStyle); err != nil {
		return err
	}

	currentRow += 2

	// Add some summary data
	summaryData := [][]string{
		{"Total Records", content.GenerateNumber(100, 1000)},
		{"Average Value", content.GenerateCurrency()},
		{"Success Rate", content.GeneratePercentage()},
		{"Last Updated", content.GenerateDate()},
	}

	for i, row := range summaryData {
		labelCell := fmt.Sprintf("A%d", currentRow+i)
		valueCell := fmt.Sprintf("B%d", currentRow+i)

		if err := f.SetCellValue(sheetName, labelCell, row[0]); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, valueCell, row[1]); err != nil {
			return err
		}
	}

	return nil
}
