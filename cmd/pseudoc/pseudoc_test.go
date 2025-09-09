package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestMainHelp(t *testing.T) {
	// Redirect standard output to capture the printed message
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Ensure stdout is restored even if the test panics
	defer func() {
		os.Stdout = oldStdout
	}()

	// Call the run function with help command
	err := run([]string{"pseudoc", "help"})
	if err != nil {
		t.Fatalf("Expected no error for help command, got: %v", err)
	}

	// Close the writer end of the pipe to flush buffers
	w.Close()

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify the output contains expected help text
	if !strings.Contains(output, "pseudoc - Lorem Ipsum for Documents") {
		t.Errorf("Help output doesn't contain expected text. Got: %s", output)
	}

	// Check for examples section
	if !strings.Contains(output, "EXAMPLES:") {
		t.Errorf("Help output should contain examples section. Got: %s", output)
	}

	// Check for quiet and dry-run flags
	if !strings.Contains(output, "--quiet") {
		t.Errorf("Help output should contain --quiet flag. Got: %s", output)
	}

	if !strings.Contains(output, "--dry-run") {
		t.Errorf("Help output should contain --dry-run flag. Got: %s", output)
	}
}

func TestMainVersion(t *testing.T) {
	// Redirect standard output to capture the printed message
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Ensure stdout is restored even if the test panics
	defer func() {
		os.Stdout = oldStdout
	}()

	// Call the run function with version command
	err := run([]string{"pseudoc", "version"})
	if err != nil {
		t.Fatalf("Expected no error for version command, got: %v", err)
	}

	// Close the writer end of the pipe to flush buffers
	w.Close()

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify the output contains version information
	if !strings.Contains(output, "pseudoc version") {
		t.Errorf("Version output doesn't contain expected text. Got: %s", output)
	}

	// Check for enhanced version info
	if !strings.Contains(output, "Go version:") {
		t.Errorf("Version output should contain Go version. Got: %s", output)
	}

	if !strings.Contains(output, "Platform:") {
		t.Errorf("Version output should contain platform info. Got: %s", output)
	}
}

func TestInvalidDocType(t *testing.T) {
	// Call the run function with an invalid document type
	err := run([]string{"pseudoc", "invalid-type"})

	// Verify we get an error
	if err == nil {
		t.Fatalf("Expected error for invalid document type, got none")
	}

	// Verify the error message mentions the invalid document type and provides help
	if !strings.Contains(err.Error(), "unknown document type") {
		t.Errorf("Expected error to mention unknown document type. Got: %v", err)
	}

	// Check that the error provides helpful information
	if !strings.Contains(err.Error(), "Available types:") {
		t.Errorf("Expected error to list available types. Got: %v", err)
	}
}

func TestDryRunMode(t *testing.T) {
	// Redirect standard output to capture the printed message
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		os.Stdout = oldStdout
	}()

	// Test dry run mode
	err := run([]string{"pseudoc", "pdf", "--dry-run", "--count", "2"})
	if err != nil {
		t.Fatalf("Expected no error for dry run, got: %v", err)
	}

	w.Close()

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify dry run output
	if !strings.Contains(output, "DRY RUN MODE") {
		t.Errorf("Dry run output should contain 'DRY RUN MODE'. Got: %s", output)
	}

	if !strings.Contains(output, "Files that would be generated:") {
		t.Errorf("Dry run output should list files to be generated. Got: %s", output)
	}
}

func TestQuietMode(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Redirect standard output to capture the printed message
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		os.Stdout = oldStdout
	}()

	// Test quiet mode - should generate file but produce minimal output
	err := run([]string{"pseudoc", "pdf", "--quiet", "--output-dir", tempDir})
	if err != nil {
		t.Fatalf("Expected no error in quiet mode, got: %v", err)
	}

	w.Close()

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// In quiet mode, there should be minimal to no output
	if strings.Contains(output, "Generating") {
		t.Errorf("Quiet mode should suppress progress messages. Got: %s", output)
	}
}

func TestInvalidFlags(t *testing.T) {
	// Test with invalid flag
	err := run([]string{"pseudoc", "pdf", "--invalid-flag"})

	if err == nil {
		t.Fatalf("Expected error for invalid flag, got none")
	}

	// Check that error provides helpful information
	if !strings.Contains(err.Error(), "invalid flag") {
		t.Errorf("Expected error to mention invalid flag. Got: %v", err)
	}
}

func TestEnhancedErrorMessages(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		errorMsg string
	}{
		{
			name:     "no command",
			args:     []string{"pseudoc"},
			errorMsg: "no command specified",
		},
		{
			name:     "invalid count",
			args:     []string{"pseudoc", "pdf", "--count", "0"},
			errorMsg: "count must be at least 1",
		},
		{
			name:     "invalid pages",
			args:     []string{"pseudoc", "pdf", "--pages", "0"},
			errorMsg: "pages must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(tt.args)
			if err == nil {
				t.Fatalf("Expected error for %s, got none", tt.name)
			}

			if !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error to contain '%s', got: %v", tt.errorMsg, err)
			}
		})
	}
}

func TestRandomDocumentType(t *testing.T) {
	// Helper function to capture output
	captureOutput := func(args []string) (string, error) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := run(args)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		return buf.String(), err
	}

	// Test with fixed seed to ensure reproducibility
	output1, err1 := captureOutput([]string{"pseudoc", "random", "--seed", "42", "--dry-run"})
	if err1 != nil {
		t.Fatalf("Expected no error for random type with seed, got: %v", err1)
	}

	// Run again with same seed and verify it's deterministic
	output2, err2 := captureOutput([]string{"pseudoc", "random", "--seed", "42", "--dry-run"})
	if err2 != nil {
		t.Fatalf("Expected no error for second random type with seed, got: %v", err2)
	}

	// The outputs should be identical when using the same seed
	if output1 != output2 {
		t.Errorf("Expected identical output with same seed.\nFirst run:\n%s\nSecond run:\n%s", output1, output2)
	}

	// Verify that the output contains expected dry-run elements
	if !strings.Contains(output1, "DRY RUN MODE") {
		t.Errorf("Expected dry-run output to contain 'DRY RUN MODE'. Got: %s", output1)
	}

	if !strings.Contains(output1, "Randomly selected document type:") {
		t.Errorf("Expected output to show randomly selected type. Got: %s", output1)
	}

	// Test that different seeds produce different results
	output3, err3 := captureOutput([]string{"pseudoc", "random", "--seed", "123", "--dry-run"})
	if err3 != nil {
		t.Fatalf("Expected no error for random type with different seed, got: %v", err3)
	}

	// With different seeds, we might get different document types
	// NOTE: This test might occasionally fail if both seeds happen to generate the same type
	// In practice, this is very unlikely with different seeds
	if output1 == output3 {
		t.Logf("Warning: Different seeds produced identical output, which is statistically unlikely but possible")
	}
}

func TestCommandAliases(t *testing.T) {
	aliases := [][]string{
		{"pseudoc", "word", "--dry-run"},
		{"pseudoc", "docx", "--dry-run"},
		{"pseudoc", "spreadsheet", "--dry-run"},
		{"pseudoc", "sheet", "--dry-run"},
		{"pseudoc", "xlsx", "--dry-run"},
	}

	for _, args := range aliases {
		err := run(args)
		if err != nil {
			t.Errorf("Expected no error for alias %v, got: %v", args[1], err)
		}
	}
}

func TestVersionAliases(t *testing.T) {
	aliases := []string{"-v", "--version", "version"}

	for _, alias := range aliases {
		// Capture output for each alias test
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := run([]string{"pseudoc", alias})

		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("Expected no error for version alias %s, got: %v", alias, err)
			continue
		}

		// Read and validate the output
		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		if !strings.Contains(output, "pseudoc version") {
			t.Errorf("Version alias %s should produce version output, got: %s", alias, output)
		}
	}
}
