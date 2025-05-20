package converter

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/JerkyTreats/PHITE/converter/internal/models"
)

func TestSaveResult(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "result_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Create a sample ConversionResult
	result := &models.ConversionResult{
		Groupings: []models.Grouping{
			{
				Topic: "Test Topic",
				Name:  "Test Group",
				SNP: []models.SNP{
					{
						Gene:   "TestGene",
						RSID:   "rs123",
						Allele: "A",
						Notes:  "Test SNP",
						Subject: models.Subject{
							Genotype: "AA",
							Match:    "Full",
						},
					},
				},
			},
		},
	}

	// Test saving
	err = SaveResult(result, tempFile.Name())
	if err != nil {
		t.Fatalf("SaveResult failed: %v", err)
	}

	// Verify file exists and has content
	stat, err := os.Stat(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if stat.Size() == 0 {
		t.Fatal("Saved file is empty")
	}

	// Verify JSON content
	content, err := os.ReadFile(tempFile.Name())
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

	parser := NewTSVParser(testFile)
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify we got one grouping
	if len(result.Result.Groupings) != 1 {
		t.Errorf("Expected 1 grouping, got %d", len(result.Result.Groupings))
	}

	// Verify the grouping details
	grouping := result.Result.Groupings[0]
	if grouping.Topic != "Nutrients – Vitamins and Minerals" {
		t.Errorf("Expected topic 'Nutrients – Vitamins and Minerals', got %s", grouping.Topic)
	}
	if grouping.Name != "MTHFR" {
		t.Errorf("Expected name 'MTHFR', got %s", grouping.Name)
	}

	// Verify we got 4 SNPs (one was skipped due to "--" genotype)
	if len(grouping.SNP) != 4 {
		t.Errorf("Expected 4 SNPs, got %d", len(grouping.SNP))
	}

	// Verify SNP details
	snp := grouping.SNP[0]
	if snp.Gene != "MTHFR C677T" {
		t.Errorf("Expected gene 'MTHFR C677T', got %s", snp.Gene)
	}
	if snp.RSID != "rs1801133" {
		t.Errorf("Expected RSID 'rs1801133', got %s", snp.RSID)
	}
	if snp.Allele != "A" {
		t.Errorf("Expected allele 'A', got %s", snp.Allele)
	}
	if snp.Subject.Genotype != "AG" {
		t.Errorf("Expected genotype 'AG', got %s", snp.Subject.Genotype)
	}
	if snp.Subject.Match != "Partial" {
		t.Errorf("Expected match 'Partial', got %s", snp.Subject.Match)
	}

	// Verify full match SNP
	fullMatchSNP := grouping.SNP[3]
	if fullMatchSNP.Subject.Genotype != "AA" {
		t.Errorf("Expected genotype 'AA', got %s", fullMatchSNP.Subject.Genotype)
	}
	if fullMatchSNP.Subject.Match != "Full" {
		t.Errorf("Expected match 'Full', got %s", fullMatchSNP.Subject.Match)
	}

	// Verify we got error records for skipped records
	if len(result.ErrorRecords) != 1 {
		t.Errorf("Expected 1 error record, got %d", len(result.ErrorRecords))
	}
	if !strings.Contains(result.ErrorRecords[0], "Record with special genotype '--'") {
		t.Errorf("Expected error record to contain special genotype message, got: %s", result.ErrorRecords[0])
	}
}

func TestParseInvalidTSV(t *testing.T) {
	// Create a temporary invalid TSV file
	tempFile, err := os.CreateTemp("", "invalid_*.tsv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write invalid TSV data (wrong number of columns)
	testRecord := "Foo\tBar\tBaz\tQux\tQuux\tCorge"
	_, err = tempFile.WriteString("Topic\tGroup\tGene\tRS ID\tAllele\tSubject Genotype\n" + testRecord)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	parser := NewTSVParser(tempFile.Name())
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify we got the error record
	if len(result.ErrorRecords) != 1 {
		t.Errorf("Expected 1 error record, got %d", len(result.ErrorRecords))
	}
	if !strings.Contains(result.ErrorRecords[0], "Record with 6 columns") {
		t.Errorf("Expected error record to contain column count, got: %s", result.ErrorRecords[0])
	}
	// Check that each field is present in the error record
	fields := []string{"Foo", "Bar", "Baz", "Qux", "Quux", "Corge"}
	for _, field := range fields {
		if !strings.Contains(result.ErrorRecords[0], field) {
			t.Errorf("Expected error record to contain field %q, got: %s", field, result.ErrorRecords[0])
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

	parser := NewTSVParser(tempFile.Name())
	_, err = parser.Parse()
	if err == nil {
		t.Error("Expected error for empty file")
	}
}

func TestParseMissingFile(t *testing.T) {
	parser := NewTSVParser("/nonexistent/file.tsv")
	_, err := parser.Parse()
	if err == nil {
		t.Error("Expected error for missing file")
	}
}
