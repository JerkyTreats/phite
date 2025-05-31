package genotype_test

import (
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"testing"

	"phite.io/polygenic-risk-calculator/internal/genotype"
	"phite.io/polygenic-risk-calculator/internal/model"
	"phite.io/polygenic-risk-calculator/internal/logging"
)

// Helper function to sort slices in ParseGenotypeDataOutput for consistent comparison
func sortOutput(output *genotype.ParseGenotypeDataOutput) {
	sort.Slice(output.UserGenotypes, func(i, j int) bool { return output.UserGenotypes[i].RSID < output.UserGenotypes[j].RSID })
	sort.Slice(output.ValidatedSNPs, func(i, j int) bool { return output.ValidatedSNPs[i].RSID < output.ValidatedSNPs[j].RSID })
	sort.Strings(output.SNPsMissing)
}

func TestParseGenotypeData(t *testing.T) {
	logging.SetSilentLoggingForTest()
	// Define a common set of requested RSIDs and mock GWAS data for tests
	baseRequestedRSIDs := []string{"rs1001", "rs1002", "rs1003", "rs1004", "rs2001", "rs2002", "rs9999"}
	mockGWASData := map[string]model.GWASSNPRecord{
		"rs1001": {RSID: "rs1001", RiskAllele: "A"},
		"rs1002": {RSID: "rs1002", RiskAllele: "C"},
		// rs1003 not in GWAS mock
		"rs2001": {RSID: "rs2001", RiskAllele: "G"},
		// rs2002 not in GWAS mock
		// rs1004 not in GWAS mock
	}

	// Attempt to construct absolute paths to test data files
	// This assumes tests might be run from the project root or the package directory.
	// Adjust if your test execution context is different.
	_, currentTestFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("Failed to get current test file path")
	}
	testDataDir := filepath.Join(filepath.Dir(currentTestFile), "testdata")

	tests := []struct {
		name       string
		inputFile  string // Relative to testDataDir
		input      genotype.ParseGenotypeDataInput
		wantOutput genotype.ParseGenotypeDataOutput
		wantErr    bool
	}{
		{
			name:      "Valid AncestryDNA file",
			inputFile: "ancestry_valid.txt",
			input: genotype.ParseGenotypeDataInput{
				RequestedRSIDs: baseRequestedRSIDs,
				GWASData:       mockGWASData,
			},
			wantOutput: genotype.ParseGenotypeDataOutput{
				UserGenotypes: []model.UserGenotype{
					{RSID: "rs1001", Genotype: "AG"},
					{RSID: "rs1002", Genotype: "CC"},
					{RSID: "rs1003", Genotype: "TA"},
				},
				ValidatedSNPs: []model.ValidatedSNP{
					{RSID: "rs1001", Genotype: "AG", FoundInGWAS: true},
					{RSID: "rs1002", Genotype: "CC", FoundInGWAS: true},
					{RSID: "rs1003", Genotype: "TA", FoundInGWAS: false},
				},
				SNPsMissing: []string{"rs1004", "rs2001", "rs2002", "rs9999"}, // rs1004 (non-GACT), others not in file
			},
			wantErr: false,
		},
		{
			name:      "Valid 23andMe file",
			inputFile: "23andme_valid.txt",
			input: genotype.ParseGenotypeDataInput{
				RequestedRSIDs: baseRequestedRSIDs, // Use a relevant subset or all
				GWASData:       mockGWASData,
			},
			wantOutput: genotype.ParseGenotypeDataOutput{
				UserGenotypes: []model.UserGenotype{
					{RSID: "rs2001", Genotype: "AG"},
					{RSID: "rs2002", Genotype: "CC"},
					// rs2003 is TA, not in baseRequestedRSIDs for this example, but could be added
				},
				ValidatedSNPs: []model.ValidatedSNP{
					{RSID: "rs2001", Genotype: "AG", FoundInGWAS: true},
					{RSID: "rs2002", Genotype: "CC", FoundInGWAS: false}, // Not in mockGWASData
				},
				// rs1001, rs1002, rs1003, rs1004 are from AncestryDNA test, not expected here
				// rs2004 (non-GACT), rs2005 (non-GACT)
				// rs9999 (not in file)
				SNPsMissing: []string{"rs1001", "rs1002", "rs1003", "rs1004", "rs9999"}, // rs2004, rs2005 if requested
			},
			wantErr: false,
		},
		{
			name:      "Empty file",
			inputFile: "empty.txt",
			input: genotype.ParseGenotypeDataInput{
				RequestedRSIDs: baseRequestedRSIDs,
				GWASData:       mockGWASData,
			},
			wantOutput: genotype.ParseGenotypeDataOutput{
				UserGenotypes: nil,                // Or []model.UserGenotype{}
				ValidatedSNPs: nil,                // Or []model.ValidatedSNP{}
				SNPsMissing:   baseRequestedRSIDs, // All requested SNPs will be missing
			},
			wantErr: false, // Empty file is not an error, just no data found
		},
		{
			name:      "File not found",
			inputFile: "non_existent_file.txt", // This file should not exist
			input: genotype.ParseGenotypeDataInput{
				RequestedRSIDs: baseRequestedRSIDs,
				GWASData:       mockGWASData,
			},
			wantOutput: genotype.ParseGenotypeDataOutput{}, // Zero-value struct
			wantErr:    true,
		},
		// TODO: Add more test cases:
		// - Header-only AncestryDNA
		// - Header-only 23andMe
		// - Malformed AncestryDNA (should skip bad lines, process good ones)
		// - SNPs with various non-GACT genotypes (e.g., "N", "00", "G", "-") if requested
		// - Case sensitivity of RSIDs (e.g. "rs1001" vs "RS1001") - spec implies case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentInput := tt.input
			currentInput.GenotypeFilePath = filepath.Join(testDataDir, tt.inputFile)

			gotOutput, err := genotype.ParseGenotypeData(currentInput)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGenotypeData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Sort slices for consistent comparison before DeepEqual
			sortOutput(&gotOutput)
			sortOutput(&tt.wantOutput)

			if !reflect.DeepEqual(gotOutput.UserGenotypes, tt.wantOutput.UserGenotypes) {
				t.Errorf("ParseGenotypeData() got UserGenotypes = %v, want %v", gotOutput.UserGenotypes, tt.wantOutput.UserGenotypes)
			}
			if !reflect.DeepEqual(gotOutput.ValidatedSNPs, tt.wantOutput.ValidatedSNPs) {
				t.Errorf("ParseGenotypeData() got ValidatedSNPs = %v, want %v", gotOutput.ValidatedSNPs, tt.wantOutput.ValidatedSNPs)
			}
			if !reflect.DeepEqual(gotOutput.SNPsMissing, tt.wantOutput.SNPsMissing) {
				t.Errorf("ParseGenotypeData() got SNPsMissing = %v, want %v", gotOutput.SNPsMissing, tt.wantOutput.SNPsMissing)
			}
		})
	}
}
