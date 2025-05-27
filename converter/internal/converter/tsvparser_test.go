package converter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
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
	parser := NewTSVParser(tempFile.Name(), tempDir, "group")
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

	parser := NewTSVParser(testFile, tempDir, "group")
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

	parser := NewTSVParser(tempFile.Name(), tempDir, "group")
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

	parser := NewTSVParser(tempFile.Name(), tempDir, "group")
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

	parser := NewTSVParser("/nonexistent/file.tsv", tempDir, "group")
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

func TestParseGroupByTopic(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_output_topic_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tsvData := "Topic\tGroup\tGene\tRS ID\tAllele\tSubject Genotype\tNotes\n"
	tsvData += "Alpha\tA1\tGene1\trs1\tA\tAA\tNote1\n" // SNP for Alpha, Group A1
	tsvData += "Beta\tB1\tGene2\trs2\tG\tGG\tNote2\n"  // SNP for Beta, Group B1
	tsvData += "Alpha\tA2\tGene3\trs3\tT\tTT\tNote3\n" // SNP for Alpha, Group A2

	tempFile, err := os.CreateTemp(tempDir, "test_topic_*.tsv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	filePath := tempFile.Name()
	_, err = tempFile.WriteString(tsvData)
	if err != nil {
		t.Fatalf("Failed to write test TSV: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	parser := NewTSVParser(filePath, tempDir, "topic")
	outputFiles, _, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(outputFiles) != 2 {
		t.Fatalf("Expected 2 output files (one per topic Alpha, Beta), got %d. Files: %v", len(outputFiles), outputFiles)
	}

	type ExpectedSubject struct {
		Genotype string `json:"Genotype"`
		Match    string `json:"Match"`
	}

	type ExpectedSNP struct {
		Gene    string          `json:"Gene"`
		RSID    string          `json:"RSID"`
		Allele  string          `json:"Allele"`
		Notes   string          `json:"Notes"`
		Subject ExpectedSubject `json:"Subject"`
	}

	type TopicOutput struct {
		Topic     string                  `json:"Topic"`
		Groupings map[string][]ExpectedSNP `json:"Groupings"`
	}

	expectedOutputs := map[string]TopicOutput{
		"Alpha.json": {
			Topic: "Alpha",
			Groupings: map[string][]ExpectedSNP{
				"A1": {{Gene: "Gene1", RSID: "rs1", Allele: "A", Notes: "Note1", Subject: ExpectedSubject{Genotype: "AA", Match: "Full"}}},
				"A2": {{Gene: "Gene3", RSID: "rs3", Allele: "T", Notes: "Note3", Subject: ExpectedSubject{Genotype: "TT", Match: "Full"}}},
			},
		},
		"Beta.json": {
			Topic: "Beta",
			Groupings: map[string][]ExpectedSNP{
				"B1": {{Gene: "Gene2", RSID: "rs2", Allele: "G", Notes: "Note2", Subject: ExpectedSubject{Genotype: "GG", Match: "Full"}}},
			},
		},
	}

	for _, f := range outputFiles {
		fileName := filepath.Base(f)
		expected, ok := expectedOutputs[fileName]
		if !ok {
			t.Errorf("Unexpected output file: %s", fileName)
			continue
		}

		content, err := os.ReadFile(f)
		if err != nil {
			t.Errorf("Failed to read output file %s: %v", fileName, err)
			continue
		}

		var actual TopicOutput
		if err := json.Unmarshal(content, &actual); err != nil {
			t.Errorf("Failed to unmarshal JSON from %s: %v. Content:\n%s", fileName, err, string(content))
			continue
		}

		if actual.Topic != expected.Topic {
			t.Errorf("File %s: Expected Topic '%s', got '%s'", fileName, expected.Topic, actual.Topic)
		}

		if !reflect.DeepEqual(actual.Groupings, expected.Groupings) {
			t.Errorf("File %s: Groupings mismatch.\nExpected: %+v\nActual:   %+v", fileName, expected.Groupings, actual.Groupings)
			// For more detailed diff:
			expectedJSON, _ := json.MarshalIndent(expected.Groupings, "", "  ")
			actualJSON, _ := json.MarshalIndent(actual.Groupings, "", "  ")
			t.Logf("File %s: Expected Groupings JSON:\n%s", fileName, string(expectedJSON))
			t.Logf("File %s: Actual Groupings JSON:\n%s", fileName, string(actualJSON))
		}
	}
}

func TestParseWithMixedTopicsAndGroups(t *testing.T) {
	// RED: This test will fail until grouping flexibility is implemented
	_, curFilename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Could not get caller information")
	}
	testDir := filepath.Dir(curFilename)
	testFile := filepath.Join(testDir, "testdata", "sample.tsv")

	// Test group mode
	tempDirGroup, err := os.MkdirTemp("", "test_output_mixed_group_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory for group test: %v", err)
	}
	defer os.RemoveAll(tempDirGroup)

	parserGroup := NewTSVParser(testFile, tempDirGroup, "group")
	filesGroup, _, err := parserGroup.Parse()
	if err != nil {
		t.Fatalf("Parse (group mode) failed: %v", err)
	}
	if len(filesGroup) != 1 {
		t.Errorf("Expected 1 output file for group mode with sample.tsv, got %d. Files: %v", len(filesGroup), filesGroup)
	}
	if filepath.Base(filesGroup[0]) != "MTHFR.json" {
		t.Errorf("Expected file MTHFR.json for group mode, got %s", filepath.Base(filesGroup[0]))
	}

	// Test topic mode
	tempDirTopic, err := os.MkdirTemp("", "test_output_mixed_topic_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory for topic test: %v", err)
	}
	defer os.RemoveAll(tempDirTopic)

	parserTopic := NewTSVParser(testFile, tempDirTopic, "topic")
	filesTopic, _, err := parserTopic.Parse()
	if err != nil {
		t.Fatalf("Parse (topic mode) failed: %v", err)
	}
	if len(filesTopic) != 1 {
		t.Errorf("Expected 1 output file for topic mode with sample.tsv, got %d. Files: %v", len(filesTopic), filesTopic)
	}

	expectedTopicFilename := "Nutrients – Vitamins and Minerals.json" // Based on current simple ReplaceAll("/", "-")
	if filepath.Base(filesTopic[0]) != expectedTopicFilename {
		t.Errorf("Expected file '%s' for topic mode, got '%s'", expectedTopicFilename, filepath.Base(filesTopic[0]))
	}
}

func TestInvalidGroupingMode(t *testing.T) {
	// RED: This test will fail until error handling for invalid mode is implemented
	tempDir, err := os.MkdirTemp("", "test_output_invalidmode_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tsvData := "Topic\tGroup\tGene\tRS ID\tAllele\tSubject Genotype\tNotes\nAlpha\tA1\tGene1\trs1\tA\tAA\tNote1\n"
	tempFile, err := os.CreateTemp(tempDir, "test_invalidmode_*.tsv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	filePath := tempFile.Name()
	_, err = tempFile.WriteString(tsvData)
	if err != nil {
		t.Fatalf("Failed to write test TSV: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	parser := NewTSVParser(filePath, tempDir, "invalidmode")
	_, _, err = parser.Parse()
	if err == nil {
		t.Error("Expected error for invalid grouping mode, got nil")
	} else {
		if !strings.Contains(strings.ToLower(err.Error()), "invalid grouping mode") &&
			!strings.Contains(strings.ToLower(err.Error()), "unknown grouping mode") {
			t.Errorf("Expected error message to relate to invalid/unknown grouping mode, got: %v", err)
		}
	}
}
