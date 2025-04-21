package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

// TODO: figure out how to automagically update this whenenever I run `just release`
const version = "0.1.0"

// Supported Document Types
const (
	DocTypePDF  = "pdf"
	DocTypeDOCX = "docx"
	DocTypeXLSX = "xlsx"
)

// CLI Options
type Options struct {
	OutputDir string
	Filename  string
	Count     int
	Pages     int
	Sheets    int
	Size      string
	Seed      int64
}

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

	// Determine document type from command
	docType, err := parseDocumentType(command)
	if err != nil {
		return err
	}

	// parse the remaining options
	options, err := parseOptions(args[2:])
	if err != nil {
		return err
	}

	return executeCommand(docType, options)
}

func parseDocumentType(command string) (string, error) {
	switch command {
	case "pdf":
		return DocTypePDF, nil
	case "docx", "word":
		return DocTypeDOCX, nil
	case "xlsx", "spreadsheet", "sheet":
		return DocTypeXLSX, nil
	case "random":
		// for now, let's just return pdf
		return DocTypePDF, nil
	default:
		return "", fmt.Errorf("unknown document type: %s", command)
	}
}

func parseOptions(args []string) (Options, error) {
	// Create a new FlagSet
	flagSet := flag.NewFlagSet("options", flag.ContinueOnError)

	// Let's define the options
	options := Options{}
	flagSet.StringVar(&options.OutputDir, "output-dir", ".", "Directory to store output files")
	flagSet.StringVar(&options.Filename, "filename", "", "Base filename (will append format extension)")
	flagSet.IntVar(&options.Count, "count", 1, "Number of documents to generate")
	flagSet.IntVar(&options.Pages, "pages", 1, "For DOCX/PDF: Number of pages")
	flagSet.IntVar(&options.Sheets, "sheets", 1, "For XLSX: Number of sheets")
	flagSet.StringVar(&options.Size, "size", "", "Target file size (supports KB, MB, GB)")
	flagSet.Int64Var(&options.Seed, "seed", 0, "Set random seed for reproducble output")

	// Parse the flags
	if err := flagSet.Parse(args); err != nil {
		return Options{}, err
	}

	// Validate options
	if options.Count < 1 {
		return Options{}, errors.New("count must be at least 1")
	}

	fmt.Println("------ Debugging -------")
	fmt.Println("Here are the selected options:")
	fmt.Printf("%#v\n", options)
	fmt.Println("------------------------")
	return options, nil
}

func executeCommand(docType string, options Options) error {
	// For now, we print some statements
	fmt.Printf("Generating %d %s document(s)\n", options.Count, docType)
	fmt.Printf("Output directory: %s\n", options.OutputDir)
	if options.Filename != "" {
		fmt.Printf("Base filename: %s\n", options.Filename)
	}

	fmt.Println("Document generation not yet implemented")
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
