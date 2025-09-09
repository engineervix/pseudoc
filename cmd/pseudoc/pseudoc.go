package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/engineervix/pseudoc/internal/config"
	"github.com/engineervix/pseudoc/internal/generator"
)

// TODO: figure out how to automagically update this whenever I run `just release`
const version = "0.1.0"

// Build information - can be set via ldflags during build
var (
	buildTime = "dev"
	gitCommit = "dev"
)

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
		printVersionInfo()
		return nil
	}

	// Initialize config with defaults
	cfg := config.DefaultConfig()

	// parse the remaining options first to get seed
	if err := parseOptions(args[2:], &cfg); err != nil {
		return enhanceError(err, "parsing command options")
	}

	// Create random number generator
	var rng *rand.Rand
	if cfg.Seed != 0 {
		rng = rand.New(rand.NewSource(cfg.Seed))
	} else {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	// Determine document type from command using the seeded generator
	docType, err := parseDocumentType(command, rng)
	if err != nil {
		return enhanceError(err, "parsing document type")
	}
	cfg.DocType = docType

	// Track if random was used for user feedback
	isRandom := command == "random"

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return enhanceError(err, "validating configuration")
	}

	return generateDocuments(&cfg, isRandom, command)
}

func parseDocumentType(command string, rng *rand.Rand) (string, error) {
	switch command {
	case "pdf":
		return config.DocTypePDF, nil
	case "docx", "word":
		return config.DocTypeDOCX, nil
	case "xlsx", "spreadsheet", "sheet":
		return config.DocTypeXLSX, nil
	case "random":
		// Randomly select one of the available document types
		types := []string{config.DocTypePDF, config.DocTypeDOCX, config.DocTypeXLSX}
		return types[rng.Intn(len(types))], nil
	default:
		return "", fmt.Errorf("unknown document type: %s\n\nAvailable types: pdf, docx, word, xlsx, spreadsheet, sheet, random\nUse 'pseudoc help' for more information", command)
	}
}

func parseOptions(args []string, cfg *config.Config) error {
	// Create a new FlagSet with custom error handling
	flagSet := flag.NewFlagSet("options", flag.ContinueOnError)

	// Suppress default error output since we'll provide our own
	flagSet.Usage = func() {}

	// Let's define the options
	flagSet.StringVar(&cfg.OutputDir, "output-dir", ".", "Directory to store output files")
	flagSet.StringVar(&cfg.Filename, "filename", "", "Base filename (will append format extension)")
	flagSet.IntVar(&cfg.Count, "count", 1, "Number of documents to generate")
	flagSet.IntVar(&cfg.Pages, "pages", 1, "For DOCX/PDF: Number of pages")
	flagSet.IntVar(&cfg.Sheets, "sheets", 1, "For XLSX: Number of sheets")
	flagSet.Int64Var(&cfg.Seed, "seed", 0, "Set random seed for reproducible output")
	flagSet.BoolVar(&cfg.Quiet, "quiet", false, "Suppress output for scripting")
	flagSet.BoolVar(&cfg.DryRun, "dry-run", false, "Preview what would be generated without creating files")

	// Parse the flags
	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("invalid flag: %w\n\nUse 'pseudoc help' to see available options", err)
	}

	return nil
}

func generateDocuments(cfg *config.Config, selectedRandomly bool, originalCommand string) error {
	// Get the generator for this document type
	gen, err := generator.GetGenerator(cfg.DocType)
	if err != nil {
		return enhanceError(err, "getting document generator")
	}

	// Handle dry-run mode
	if cfg.DryRun {
		return handleDryRun(cfg, selectedRandomly)
	}

	// Show what we're generating (unless quiet)
	if !cfg.Quiet {
		if selectedRandomly {
			fmt.Printf("Randomly selected: %s\n", cfg.DocType)
		}
		fmt.Printf("Generating %d %s document(s)\n", cfg.Count, strings.ToUpper(cfg.DocType))
	}

	// Generate the requested number of documents
	for i := 0; i < cfg.Count; i++ {
		// Get the output filename
		filename := cfg.GenerateFileName(i)

		// Show progress (unless quiet)
		if !cfg.Quiet {
			if cfg.Count > 1 {
				fmt.Printf("[%d/%d] Creating %s...\n", i+1, cfg.Count, filename)
			} else {
				fmt.Printf("Creating %s...\n", filename)
			}
		}

		// Create the output file
		file, err := os.Create(filename)
		if err != nil {
			return enhanceFileError(err, filename, "create")
		}

		// Generate the document
		if err := gen.Generate(file, cfg); err != nil {
			file.Close()
			return fmt.Errorf("document generation failed for %s: %w", filename, err)
		}

		file.Close()
	}

	if !cfg.Quiet {
		if cfg.Count > 1 {
			fmt.Printf("✓ Successfully generated %d documents\n", cfg.Count)
		} else {
			fmt.Printf("✓ Document generated successfully\n")
		}
	}

	return nil
}

func handleDryRun(cfg *config.Config, selectedRandomly bool) error {
	fmt.Println("DRY RUN MODE - No files will be created")
	fmt.Println("=======================================")

	if selectedRandomly {
		fmt.Printf("Randomly selected document type: %s\n", strings.ToUpper(cfg.DocType))
	} else {
		fmt.Printf("Document type: %s\n", strings.ToUpper(cfg.DocType))
	}

	fmt.Printf("Output directory: %s\n", cfg.OutputDir)
	fmt.Printf("Number of documents: %d\n", cfg.Count)

	switch cfg.DocType {
	case config.DocTypePDF, config.DocTypeDOCX:
		fmt.Printf("Pages per document: %d\n", cfg.Pages)
	case config.DocTypeXLSX:
		fmt.Printf("Sheets per document: %d\n", cfg.Sheets)
	}

	if cfg.Seed != 0 {
		fmt.Printf("Random seed: %d\n", cfg.Seed)
	}

	fmt.Println("\nFiles that would be generated:")
	for i := 0; i < cfg.Count; i++ {
		filename := cfg.GenerateFileName(i)
		fmt.Printf("  - %s\n", filename)
	}

	return nil
}

func enhanceError(err error, context string) error {
	switch {
	case strings.Contains(err.Error(), "permission denied"):
		return fmt.Errorf("%s: %w\n\nTip: Check file/directory permissions or try running with appropriate privileges", context, err)
	case strings.Contains(err.Error(), "no such file or directory"):
		return fmt.Errorf("%s: %w\n\nTip: Make sure the directory exists or use --output-dir to specify a valid directory", context, err)
	case strings.Contains(err.Error(), "directory"):
		return fmt.Errorf("%s: %w\n\nTip: Use --output-dir to specify where files should be created", context, err)
	default:
		return fmt.Errorf("%s: %w", context, err)
	}
}

func enhanceFileError(err error, filename, operation string) error {
	switch {
	case strings.Contains(err.Error(), "permission denied"):
		return fmt.Errorf("cannot %s file %s: permission denied\n\nTip: Check if you have write permissions to the directory", operation, filename)
	case strings.Contains(err.Error(), "no such file or directory"):
		return fmt.Errorf("cannot %s file %s: directory doesn't exist\n\nTip: Create the directory first or use --output-dir to specify an existing directory", operation, filename)
	case strings.Contains(err.Error(), "file exists"):
		return fmt.Errorf("cannot %s file %s: file already exists\n\nTip: Use --filename to specify a different name or remove the existing file", operation, filename)
	default:
		return fmt.Errorf("cannot %s file %s: %w", operation, filename, err)
	}
}

func printVersionInfo() {
	fmt.Printf("pseudoc version %s\n", version)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	if buildTime != "dev" {
		fmt.Printf("Built: %s\n", buildTime)
	}
	if gitCommit != "dev" {
		fmt.Printf("Commit: %s\n", gitCommit)
	}
}

func printUsage() {
	fmt.Println("pseudoc - Lorem Ipsum for Documents")
	fmt.Println("Generate placeholder documents for testing and development")
	fmt.Println()

	fmt.Println("USAGE:")
	fmt.Println("  pseudoc [command] [options]")
	fmt.Println()

	fmt.Println("COMMANDS:")
	fmt.Println("  pdf                          Generate PDF document")
	fmt.Println("  docx, word                   Generate Word document")
	fmt.Println("  xlsx, spreadsheet, sheet     Generate Excel spreadsheet")
	fmt.Println("  random                       Generate random document type")
	fmt.Println("  version, -v, --version       Show version information")
	fmt.Println("  help, -h, --help             Show this help message")
	fmt.Println()

	fmt.Println("OPTIONS:")
	fmt.Println("  --count N                    Number of documents to generate (default: 1)")
	fmt.Println("  --output-dir DIR             Directory to store output files (default: current directory)")
	fmt.Println("  --filename NAME              Base filename (default: auto-generated with timestamp)")
	fmt.Println("  --seed VALUE                 Set random seed for reproducible output")
	fmt.Println("  --quiet                      Suppress output for scripting")
	fmt.Println("  --dry-run                    Preview what would be generated without creating files")
	fmt.Println()

	fmt.Println("FORMAT-SPECIFIC OPTIONS:")
	fmt.Println("  --pages N                    For PDF/DOCX: Number of pages (default: 1)")
	fmt.Println("  --sheets N                   For XLSX: Number of sheets (default: 1)")
	fmt.Println()

	fmt.Println("EXAMPLES:")
	fmt.Println("  # Generate a single PDF")
	fmt.Println("  pseudoc pdf")
	fmt.Println()
	fmt.Println("  # Generate 5 Word documents with 3 pages each")
	fmt.Println("  pseudoc docx --count 5 --pages 3")
	fmt.Println()
	fmt.Println("  # Generate Excel file with custom filename")
	fmt.Println("  pseudoc xlsx --filename my-data --sheets 4")
	fmt.Println()
	fmt.Println("  # Generate 10 random documents in a specific directory")
	fmt.Println("  pseudoc random --count 10 --output-dir ./test-files")
	fmt.Println()
	fmt.Println("  # Preview what would be generated (dry run)")
	fmt.Println("  pseudoc pdf --count 3 --dry-run")
	fmt.Println()
	fmt.Println("  # Generate reproducible documents with seed")
	fmt.Println("  pseudoc random --seed 42 --count 5")
	fmt.Println()
	fmt.Println("  # Silent generation for scripts")
	fmt.Println("  pseudoc pdf --quiet --output-dir /tmp")
	fmt.Println()

	fmt.Println("For more information, visit: https://github.com/engineervix/pseudoc")
}
