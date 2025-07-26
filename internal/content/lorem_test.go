package content

import (
	"strings"
	"testing"
)

func TestGenerateWord(t *testing.T) {
	word := GenerateWord()

	if word == "" {
		t.Error("Expected non-empty word")
	}

	// Check that it's one of the expected Lorem words
	found := false
	for _, loremWord := range loremWords {
		if word == loremWord {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Generated word '%s' not found in Lorem word list", word)
	}
}

func TestGenerateSentence(t *testing.T) {
	tests := []struct {
		name     string
		numWords int
	}{
		{"default", 0},
		{"short", 3},
		{"medium", 10},
		{"long", 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sentence := GenerateSentence(tt.numWords)

			if sentence == "" {
				t.Error("Expected non-empty sentence")
			}

			if !strings.HasSuffix(sentence, ".") {
				t.Error("Expected sentence to end with period")
			}

			// Check that first character is uppercase
			if len(sentence) > 0 && sentence[0] < 'A' || sentence[0] > 'Z' {
				t.Error("Expected sentence to start with uppercase letter")
			}

			// For specific word counts, verify approximate length
			if tt.numWords > 0 {
				words := strings.Fields(strings.TrimSuffix(sentence, "."))
				if len(words) != tt.numWords {
					t.Errorf("Expected %d words, got %d", tt.numWords, len(words))
				}
			}
		})
	}
}

func TestGenerateParagraph(t *testing.T) {
	tests := []struct {
		name     string
		numWords int
	}{
		{"default", 0},
		{"short", 20},
		{"medium", 50},
		{"long", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paragraph := GenerateParagraph(tt.numWords)

			if paragraph == "" {
				t.Error("Expected non-empty paragraph")
			}

			// Should contain multiple sentences
			sentences := strings.Split(paragraph, ". ")
			if len(sentences) < 2 {
				t.Error("Expected paragraph to contain multiple sentences")
			}
		})
	}
}

func TestGenerateTitle(t *testing.T) {
	title := GenerateTitle()

	if title == "" {
		t.Error("Expected non-empty title")
	}

	// Check that it's one of the expected titles
	found := false
	for _, sampleTitle := range sampleTitles {
		if title == sampleTitle {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Generated title '%s' not found in sample titles list", title)
	}
}

func TestGenerateText(t *testing.T) {
	text := GenerateText(3, 50)

	if text == "" {
		t.Error("Expected non-empty text")
	}

	// Should contain paragraph separators
	paragraphs := strings.Split(text, "\n\n")
	if len(paragraphs) != 3 {
		t.Errorf("Expected 3 paragraphs, got %d", len(paragraphs))
	}
}

func TestGenerateTableData(t *testing.T) {
	data := GenerateTableData(3, 4)

	if len(data) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(data))
	}

	for i, row := range data {
		if len(row) != 4 {
			t.Errorf("Expected 4 columns in row %d, got %d", i, len(row))
		}

		for j, cell := range row {
			if cell == "" {
				t.Errorf("Expected non-empty cell at [%d][%d]", i, j)
			}
		}
	}
}

func TestGenerateListItems(t *testing.T) {
	items := GenerateListItems(5)

	if len(items) != 5 {
		t.Errorf("Expected 5 items, got %d", len(items))
	}

	for i, item := range items {
		if item == "" {
			t.Errorf("Expected non-empty item at index %d", i)
		}

		if !strings.HasSuffix(item, ".") {
			t.Errorf("Expected item %d to end with period", i)
		}
	}
}

func TestGenerateDate(t *testing.T) {
	date := GenerateDate()

	if date == "" {
		t.Error("Expected non-empty date")
	}

	// Check date format (YYYY-MM-DD)
	expectedLen := 10 // "2006-01-02" format
	if len(date) != expectedLen {
		t.Errorf("Expected date length %d, got %d", expectedLen, len(date))
	}

	// Simple format check - should contain two hyphens
	hyphenCount := 0
	for _, char := range date {
		if char == '-' {
			hyphenCount++
		}
	}
	if hyphenCount != 2 {
		t.Errorf("Expected 2 hyphens in date format, got %d in %s", hyphenCount, date)
	}
}

func TestGenerateNumber(t *testing.T) {
	tests := []struct {
		name string
		min  int
		max  int
	}{
		{"normal range", 10, 100},
		{"same values", 50, 50},
		{"invalid range", 100, 50}, // max < min, should be handled
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			number := GenerateNumber(tt.min, tt.max)

			if number == "" {
				t.Error("Expected non-empty number")
			}

			// Should be a valid number string (basic check)
			if len(number) == 0 {
				t.Error("Expected non-empty number string")
			}
		})
	}
}

func TestGeneratePercentage(t *testing.T) {
	percentage := GeneratePercentage()

	if percentage == "" {
		t.Error("Expected non-empty percentage")
	}

	if !strings.HasSuffix(percentage, "%") {
		t.Error("Expected percentage to end with %")
	}

	// Should contain a decimal point
	if !strings.Contains(percentage, ".") {
		t.Error("Expected percentage to contain decimal point")
	}
}

func TestGenerateCurrency(t *testing.T) {
	currency := GenerateCurrency()

	if currency == "" {
		t.Error("Expected non-empty currency")
	}

	if !strings.HasPrefix(currency, "$") {
		t.Error("Expected currency to start with $")
	}

	// Should contain a decimal point
	if !strings.Contains(currency, ".") {
		t.Error("Expected currency to contain decimal point")
	}
}
