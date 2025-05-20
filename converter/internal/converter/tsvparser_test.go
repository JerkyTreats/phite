package converter

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSaveResult(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "test_output_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a temporary test file
	tempFile, err := os.CreateTemp("", "test_*.tsv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write test data
	testData := `Topic	Group	Gene	RS ID	Allele	Subject Genotype	Notes
Test Topic	Test Group	TestGene	rs123	A	AA	Test SNP`
	_, err = tempFile.WriteString(testData)
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}

	// Create parser with temp directory
	parser := NewTSVParser(tempFile.Name(), tempDir)
	_, _, err = parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify output files exist in directory
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 output file, got %d", len(files))
	}

	// Verify the file exists and has content
	filePath := filepath.Join(tempDir, files[0].Name())
	stat, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if stat.Size() == 0 {
		t.Fatal("Saved file is empty")
	}

	// Verify JSON content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Basic JSON validation - check for expected fields
	if !strings.Contains(string(content), "Test Topic") {
		t.Errorf("Saved JSON does not contain expected topic")
	}
	if !strings.Contains(string(content), "Test Group") {
		t.Errorf("Saved JSON does not contain expected group")
	}
}

func TestParseValidTSV(t *testing.T) {
	// Get absolute path to test file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Could not get caller information")
	}
	testDir := filepath.Dir(filename)
	testFile := filepath.Join(testDir, "testdata", "sample.tsv")

	// Create a temporary directory for output
	tempDir, err := os.MkdirTemp("", "test_output_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	parser := NewTSVParser(testFile, tempDir)
	_, errorRecords, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify output files exist in directory
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 output file, got %d", len(files))
	}

	// Verify the file exists and has content
	filePath := filepath.Join(tempDir, files[0].Name())
	stat, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if stat.Size() == 0 {
		t.Fatal("Saved file is empty")
	}

	// Read the JSON content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Verify content
	if !strings.Contains(string(content), "Nutrients – Vitamins and Minerals") {
		t.Errorf("Expected topic 'Nutrients – Vitamins and Minerals' in output")
	}
	if !strings.Contains(string(content), "MTHFR") {
		t.Errorf("Expected grouping 'MTHFR' in output")
	}

	// Verify we got error records for skipped records
	if len(errorRecords) != 1 {
		t.Errorf("Expected 1 error record, got %d", len(errorRecords))
	}
	if !strings.Contains(errorRecords[0], "Record with special genotype '--'") {
		t.Errorf("Expected error record to contain special genotype message, got: %s", errorRecords[0])
	}
}

func TestParseInvalidTSV(t *testing.T) {
	// Create a temporary invalid TSV file
	tempFile, err := os.CreateTemp("", "invalid_*.tsv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Create a temporary directory for output
	tempDir, err := os.MkdirTemp("", "test_output_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Write invalid TSV data (wrong number of columns)
	testRecord := "Foo\tBar\tBaz\tQux\tQuux\tCorge"
	_, err = tempFile.WriteString("Topic\tGroup\tGene\tRS ID\tAllele\tSubject Genotype\n" + testRecord)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	parser := NewTSVParser(tempFile.Name(), tempDir)
	_, errorRecords, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify we got the error record
	if len(errorRecords) != 1 {
		t.Errorf("Expected 1 error record, got %d", len(errorRecords))
	}
	if !strings.Contains(errorRecords[0], "Record with 6 columns") {
		t.Errorf("Expected error record to contain column count, got: %s", errorRecords[0])
	}
	// Check that each field is present in the error record
	fields := []string{"Foo", "Bar", "Baz", "Qux", "Quux", "Corge"}
	for _, field := range fields {
		if !strings.Contains(errorRecords[0], field) {
			t.Errorf("Expected error record to contain field %q, got: %s", field, errorRecords[0])
		}
	}
}

func TestParseEmptyFile(t *testing.T) {
	// Create an empty file
	tempFile, err := os.CreateTemp("", "empty_*.tsv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Create a temporary directory for output
	tempDir, err := os.MkdirTemp("", "test_output_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	parser := NewTSVParser(tempFile.Name(), tempDir)
	_, _, err = parser.Parse()
	if err == nil {
		t.Error("Expected error for empty file")
	}
	// Verify no files were created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}
	if len(files) > 0 {
		t.Errorf("Expected no output files, got %d", len(files))
	}
}

func TestParseMissingFile(t *testing.T) {
	// Create a temporary directory for output
	tempDir, err := os.MkdirTemp("", "test_output_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	parser := NewTSVParser("/nonexistent/file.tsv", tempDir)
	_, _, err = parser.Parse()
	if err == nil {
		t.Error("Expected error for missing file")
	}
	// Verify no files were created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}
	if len(files) > 0 {
		t.Errorf("Expected no output files, got %d", len(files))
	}
}
