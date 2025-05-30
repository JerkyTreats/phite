package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestEntrypoint_MissingRequiredArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := RunCLI([]string{}, &stdout, &stderr)
	if exitCode == 0 {
		t.Fatalf("expected non-zero exit code for missing required arguments")
	}
	outStr := stdout.String() + stderr.String()
	if !strings.Contains(outStr, "required") || !strings.Contains(outStr, "Usage") {
		t.Errorf("expected usage or error message, got: %q", outStr)
	}
}
