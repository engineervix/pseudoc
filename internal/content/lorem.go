package content

import (
	"fmt"
	"math/rand"
	"strings"
)

// Standard Lorem Ipsum words
var loremWords = []string{
	"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit",
	"sed", "do", "eiusmod", "tempor", "incididunt", "ut", "labore", "et", "dolore",
	"magna", "aliqua", "enim", "ad", "minim", "veniam", "quis", "nostrud",
	"exercitation", "ullamco", "laboris", "nisi", "aliquip", "ex", "ea", "commodo",
	"consequat", "duis", "aute", "irure", "in", "reprehenderit", "voluptate",
	"velit", "esse", "cillum", "fugiat", "nulla", "pariatur", "excepteur", "sint",
	"occaecat", "cupidatat", "non", "proident", "sunt", "culpa", "qui", "officia",
	"deserunt", "mollit", "anim", "id", "est", "laborum", "at", "vero", "eos",
	"accusamus", "accusantium", "doloremque", "laudantium", "totam", "rem",
	"aperiam", "eaque", "ipsa", "quae", "ab", "illo", "inventore", "veritatis",
	"et", "quasi", "architecto", "beatae", "vitae", "dicta", "explicabo", "nemo",
	"ipsam", "quia", "voluptas", "aspernatur", "aut", "odit", "fugit", "sed",
	"unde", "omnis", "iste", "natus", "error", "accusantium", "doloremque",
	"laudantium", "totam", "rem", "aperiam", "eaque", "ipsa", "quae", "ab",
	"illo", "inventore", "veritatis", "et", "quasi", "architecto", "beatae",
}

// Sample titles for variety
var sampleTitles = []string{
	"Introduction to the Topic",
	"Executive Summary",
	"Analysis and Findings",
	"Methodology Overview",
	"Research Results",
	"Conclusions and Recommendations",
	"Data Analysis Report",
	"Project Overview",
	"Technical Specifications",
	"Implementation Guide",
	"Best Practices Review",
	"Performance Metrics",
	"Quality Assurance Report",
	"Strategic Planning Document",
	"Operational Guidelines",
}

// GenerateWord returns a single random Lorem Ipsum word
func GenerateWord() string {
	return loremWords[rand.Intn(len(loremWords))]
}

// GenerateSentence generates a sentence with the specified number of words
func GenerateSentence(numWords int) string {
	if numWords <= 0 {
		numWords = rand.Intn(15) + 5 // Default to 5-19 words
	}

	words := make([]string, numWords)
	for i := 0; i < numWords; i++ {
		words[i] = GenerateWord()
	}

	// Capitalize first word
	if len(words) > 0 {
		words[0] = strings.Title(words[0])
	}

	return strings.Join(words, " ") + "."
}

// GenerateParagraph generates a paragraph with the specified number of words
func GenerateParagraph(numWords int) string {
	if numWords <= 0 {
		numWords = rand.Intn(100) + 50 // Default to 50-149 words
	}

	var sentences []string
	wordsRemaining := numWords

	for wordsRemaining > 0 {
		// Generate sentences of varying lengths
		sentenceLength := rand.Intn(12) + 4 // 4-15 words per sentence
		if sentenceLength > wordsRemaining {
			sentenceLength = wordsRemaining
		}

		sentence := GenerateSentence(sentenceLength)
		sentences = append(sentences, sentence)
		wordsRemaining -= sentenceLength
	}

	return strings.Join(sentences, " ")
}

// GenerateTitle returns a random title
func GenerateTitle() string {
	return sampleTitles[rand.Intn(len(sampleTitles))]
}

// GenerateText generates Lorem Ipsum text with specified parameters
func GenerateText(paragraphs int, wordsPerParagraph int) string {
	if paragraphs <= 0 {
		paragraphs = rand.Intn(3) + 2 // Default to 2-4 paragraphs
	}

	var result []string
	for i := 0; i < paragraphs; i++ {
		paragraph := GenerateParagraph(wordsPerParagraph)
		result = append(result, paragraph)
	}

	return strings.Join(result, "\n\n")
}

// GenerateTableData generates sample data for tables
func GenerateTableData(rows, cols int) [][]string {
	if rows <= 0 {
		rows = rand.Intn(5) + 3 // Default to 3-7 rows
	}
	if cols <= 0 {
		cols = rand.Intn(3) + 2 // Default to 2-4 columns
	}

	data := make([][]string, rows)
	for i := 0; i < rows; i++ {
		data[i] = make([]string, cols)
		for j := 0; j < cols; j++ {
			// Generate different types of data based on column
			switch j % 4 {
			case 0:
				data[i][j] = GenerateWord()
			case 1:
				data[i][j] = GenerateSentence(rand.Intn(6) + 2)
			case 2:
				data[i][j] = fmt.Sprintf("$%.2f", rand.Float64()*1000)
			case 3:
				data[i][j] = fmt.Sprintf("%d%%", rand.Intn(100))
			}
		}
	}

	return data
}

// GenerateListItems generates a list of items
func GenerateListItems(count int) []string {
	if count <= 0 {
		count = rand.Intn(5) + 3 // Default to 3-7 items
	}

	items := make([]string, count)
	for i := 0; i < count; i++ {
		items[i] = GenerateSentence(rand.Intn(8) + 3) // 3-10 words per item
	}

	return items
}
