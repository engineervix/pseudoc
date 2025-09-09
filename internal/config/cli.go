package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// CLIConfig wraps the core Config with CLI-specific settings
type CLIConfig struct {
	Config // Embedded core config

	// CLI-specific options
	OutputDir string // Directory to store output files
	Filename  string // Custom filename override
	Quiet     bool   // Suppress output for scripting
	DryRun    bool   // Preview what would be generated without creating files
}

// DefaultCLIConfig returns a CLIConfig with sensible defaults
func DefaultCLIConfig() CLIConfig {
	return CLIConfig{
		Config:    DefaultConfig(),
		OutputDir: ".", // Current directory
		Quiet:     false,
		DryRun:    false,
	}
}

// ValidateCLI validates the CLI configuration including file system checks
func (c *CLIConfig) ValidateCLI() error {
	// Validate core config first
	if err := c.Config.Validate(); err != nil {
		return err
	}

	// Skip directory validation for dry-run
	if !c.DryRun {
		if err := c.validateOutputDir(); err != nil {
			return err
		}
	}

	return nil
}

// validateOutputDir checks if the output directory exists and creates it if needed
func (c *CLIConfig) validateOutputDir() error {
	// Check if directory exists
	info, err := os.Stat(c.OutputDir)
	if err == nil {
		// Path exists, check if it's a directory
		if !info.IsDir() {
			return errors.New("output path exists but isn't a directory")
		}
		return nil
	}

	// If the error isn't that the directory doesn't exist, return the error
	if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	// Create the directory
	if err := os.MkdirAll(c.OutputDir, 0755); err != nil {
		return err
	}

	return nil
}

// GenerateFileName creates a full file path for the document
func (c *CLIConfig) GenerateFileName(index int) string {
	base := c.Filename
	if base == "" {
		base = c.Config.GenerateBaseFilename()
	}

	// Add index suffix if generating multiple documents
	if c.Count > 1 {
		base += "_" + string(rune('a'+index))
	}

	// Add extension based on document type
	ext := c.Config.GetFileExtension()

	return filepath.Join(c.OutputDir, base+ext)
}

// SetBaseFilename updates the core config's BaseFilename from CLI filename
func (c *CLIConfig) SetBaseFilename() {
	if c.Filename != "" {
		c.Config.BaseFilename = c.Filename
	}
}
