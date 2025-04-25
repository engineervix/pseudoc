package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/engineervix/pseudoc/internal/config"
	"github.com/engineervix/pseudoc/internal/generator"
)

// TODO: figure out how to automagically update this whenenever I run `just release`
const version = "0.1.0"

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	// check if any arguments were provided
	if len(args) < 2 {
		printUsage()
		return errors.New("no command specified")
	}

	// Extract the command (program name is 1st argument, then the 2nd arg is command)
	command := strings.ToLower(args[1])

	// Handle help and version commands
	switch command {
	case "help", "-h", "--help":
		printUsage()
		return nil
	case "version", "-v", "--version":
		fmt.Printf("pseudoc version %s\n", version)
		return nil
	}

	// Initialize config with defaults
	cfg := config.DefaultConfig()

	// Determine document type from command
	docType, err := parseDocumentType(command)
	if err != nil {
		return err
	}
	cfg.DocType = docType

	// parse the remaining options
	if err := parseOptions(args[2:], &cfg); err != nil {
		return err
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return err
	}

	// Initialise random seed if provided
	if cfg.Seed != 0 {
		rand.New(rand.NewSource(cfg.Seed))
	}

	return generateDocuments(&cfg)
}

func parseDocumentType(command string) (string, error) {
	switch command {
	case "pdf":
		return config.DocTypePDF, nil
	case "docx", "word":
		return config.DocTypeDOCX, nil
	case "xlsx", "spreadsheet", "sheet":
		return config.DocTypeXLSX, nil
	case "random":
		// for now, let's just return pdf
		return config.DocTypePDF, nil
	default:
		return "", fmt.Errorf("unknown document type: %s", command)
	}
}

func parseOptions(args []string, cfg *config.Config) error {
	// Create a new FlagSet
	flagSet := flag.NewFlagSet("options", flag.ContinueOnError)

	// Let's define the options
	flagSet.StringVar(&cfg.OutputDir, "output-dir", ".", "Directory to store output files")
	flagSet.StringVar(&cfg.Filename, "filename", "", "Base filename (will append format extension)")
	flagSet.IntVar(&cfg.Count, "count", 1, "Number of documents to generate")
	flagSet.IntVar(&cfg.Pages, "pages", 1, "For DOCX/PDF: Number of pages")
	flagSet.IntVar(&cfg.Sheets, "sheets", 1, "For XLSX: Number of sheets")
	flagSet.StringVar(&cfg.Size, "size", "", "Target file size (supports KB, MB, GB)")
	flagSet.Int64Var(&cfg.Seed, "seed", 0, "Set random seed for reproducble output")

	// Parse the flags
	return flagSet.Parse(args)
}

func generateDocuments(cfg *config.Config) error {
	// Get the generator for this document type
	gen, err := generator.GetGenerator(cfg.DocType)
	if err != nil {
		return err
	}

	fmt.Printf("Generating %d %s document(s)\n", cfg.Count, cfg.DocType)

	// Generate the requested number of documents
	for i := 0; i < cfg.Count; i++ {
		// Get the output filename
		filename := cfg.GenerateFileName(i)
		fmt.Printf("Creating %s...\n", filename)

		// Create the output file
		file, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}

		// Generate the document
		if err := gen.Generate(file, cfg); err != nil {
			file.Close()
			return fmt.Errorf("document generation failed: %w", err)
		}

		file.Close()
	}

	fmt.Printf("Successfully generated %d document(s)\n", cfg.Count)
	return nil
}

func printUsage() {
	fmt.Println("pseudoc - Lorem Ipsum for Documents")
	fmt.Println("\nUsage:")
	fmt.Println("  pseudoc [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  pdf                          Generate PDF document")
	fmt.Println("  docx|word                    Generate Word Document")
	fmt.Println("  xlsx|spreadsheet|sheet       Generate Excel spreadsheet")
	fmt.Println("  random                       Generate random document type")
	fmt.Println("  version                      Show version information")
	fmt.Println("  help                         Show this help message")
	fmt.Println("\nGlobal options:")
	fmt.Println("  --count N                    Number of documents to generate")
	fmt.Println("  --output-dir DIR             Directory to store output files")
	fmt.Println("  --filename NAME              Base filename (will append format extension)")
	fmt.Println("  --seed VALUE                 Set random seed for reproducible output")
	fmt.Println("\nFormat-specific options:")
	fmt.Println("  --pages N                    For PDF/DOCX: Number of pages")
	fmt.Println("  --sheets N                   For XLSX: Number of sheets")
	fmt.Println("  --size VALUE                 Target file size (supports KB, MB, GB)")
}
