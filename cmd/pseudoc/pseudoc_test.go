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
}

func TestInvalidDocType(t *testing.T) {
	// Call the run function with an invalid document type
	err := run([]string{"pseudoc", "invalid-type"})

	// Verify we get an error
	if err == nil {
		t.Fatalf("Expected error for invalid document type, got none")
	}

	// Verify the error message mentions the invalid TestInvalidDocType
	if !strings.Contains(err.Error(), "unknown document type") {
		t.Errorf("Expected error to mention unknown document type. Got: %v", err)
	}
}
