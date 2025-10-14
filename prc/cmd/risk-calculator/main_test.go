package main

import (
	"bytes"
	"strings"
	"testing"

	"phite.io/polygenic-risk-calculator/internal/logging"
)

func TestEntrypoint_MissingRequiredArgs(t *testing.T) {
	logging.SetSilentLoggingForTest()
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

func TestEntrypoint_GenotypeFileNotFound(t *testing.T) {
	logging.SetSilentLoggingForTest()
	var stdout, stderr bytes.Buffer
	args := []string{"--genotype-file", "nonexistent_file.txt", "--snps", "rs123"}
	exitCode := RunCLI(args, &stdout, &stderr)
	if exitCode == 0 {
		t.Fatalf("expected non-zero exit code for missing genotype file")
	}
	if !strings.Contains(stderr.String(), "not found") {
		t.Errorf("expected error about missing genotype file, got: %q", stderr.String())
	}
}
