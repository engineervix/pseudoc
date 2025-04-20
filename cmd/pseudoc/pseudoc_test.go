package main

import (
	"bytes"
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	// Redirect standard output to capture the printed message
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Ensure stdout is restored even if the test panics
	defer func() {
		os.Stdout = oldStdout
	}()

	// Call the main function
	main()

	// Close the writer end of the pipe to flush buffers
	w.Close()

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	actual := buf.String()

	// Check if the output matches the expected value
	expected := "Hello folks!\n"
	if actual != expected {
		t.Errorf("Expected output %q but got %q", expected, actual)
	}
}
