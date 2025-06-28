package prs

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/invariance"
	"phite.io/polygenic-risk-calculator/internal/logging"
	"phite.io/polygenic-risk-calculator/internal/model"
)

func floatsAlmostEqual(a, b float64) bool {
	const eps = 1e-9
	if a > b {
		return a-b < eps
	}
	return b-a < eps
}

// setupTestConfig sets up test configuration for PRS calculator tests
func setupTestConfig(enableValidation bool, strictMode bool) {
	config.ResetForTest()
	config.SetForTest(invariance.EnableInvarianceValidationKey, enableValidation)
	config.SetForTest(invariance.StrictValidationModeKey, strictMode)
}

func TestCalculatePRS_BasicSum(t *testing.T) {
	logging.SetSilentLoggingForTest()
	setupTestConfig(false, false) // Disable validation for basic functionality test

	snps := []model.AnnotatedSNP{
		{RSID: "rs1", Genotype: "AA", RiskAllele: "A", Beta: 0.2, Dosage: 2, Trait: "trait1"},
		{RSID: "rs2", Genotype: "AG", RiskAllele: "G", Beta: -0.5, Dosage: 1, Trait: "trait2"},
		{RSID: "rs3", Genotype: "TT", RiskAllele: "T", Beta: 0.1, Dosage: 0, Trait: "trait3"},
	}

	expected := PRSResult{
		PRSScore: 2*0.2 + 1*(-0.5) + 0*0.1,
		Details: []SNPContribution{
			{"rs1", 2, 0.2, 0.4},
			{"rs2", 1, -0.5, -0.5},
			{"rs3", 0, 0.1, 0.0},
		},
	}

	result, err := CalculatePRS(snps)
	if err != nil {
		t.Fatalf("CalculatePRS returned error: %v", err)
	}

	if !floatsAlmostEqual(result.PRSScore, expected.PRSScore) {
		t.Errorf("PRSScore: got %v, want %v", result.PRSScore, expected.PRSScore)
	}
	if !reflect.DeepEqual(result.Details, expected.Details) {
		t.Errorf("Details: got %+v, want %+v", result.Details, expected.Details)
	}
}

func TestCalculatePRS_EmptyInput(t *testing.T) {
	logging.SetSilentLoggingForTest()
	setupTestConfig(false, false) // Disable validation for basic functionality test

	result, err := CalculatePRS([]model.AnnotatedSNP{})
	if err != nil {
		t.Fatalf("CalculatePRS returned error for empty input: %v", err)
	}
	if !floatsAlmostEqual(result.PRSScore, 0) {
		t.Errorf("Empty input: PRSScore = %v, want 0", result.PRSScore)
	}
	if len(result.Details) != 0 {
		t.Errorf("Empty input: Details length = %d, want 0", len(result.Details))
	}
}

func TestCalculatePRS_MissingSNPs(t *testing.T) {
	logging.SetSilentLoggingForTest()
	setupTestConfig(false, false) // Disable validation for basic functionality test

	// Simulate missing SNP by omitting from input; should just not contribute
	snps := []model.AnnotatedSNP{
		{RSID: "rs1", Genotype: "AA", RiskAllele: "A", Beta: 0.3, Dosage: 2, Trait: "trait1"},
	}

	expected := PRSResult{
		PRSScore: 0.6,
		Details: []SNPContribution{
			{"rs1", 2, 0.3, 0.6},
		},
	}

	result, err := CalculatePRS(snps)
	if err != nil {
		t.Fatalf("CalculatePRS returned error: %v", err)
	}
	if !floatsAlmostEqual(result.PRSScore, expected.PRSScore) {
		t.Errorf("PRSScore: got %v, want %v", result.PRSScore, expected.PRSScore)
	}
	if !reflect.DeepEqual(result.Details, expected.Details) {
		t.Errorf("Details: got %+v, want %+v", result.Details, expected.Details)
	}
}

func TestCalculatePRS_NegativeEffectSize(t *testing.T) {
	logging.SetSilentLoggingForTest()
	setupTestConfig(false, false) // Disable validation for basic functionality test

	snps := []model.AnnotatedSNP{
		{RSID: "rsX", Genotype: "GG", RiskAllele: "G", Beta: -1.2, Dosage: 2, Trait: "traitX"},
	}

	expected := PRSResult{
		PRSScore: 2 * -1.2,
		Details: []SNPContribution{
			{"rsX", 2, -1.2, -2.4},
		},
	}

	result, err := CalculatePRS(snps)
	if err != nil {
		t.Fatalf("CalculatePRS returned error: %v", err)
	}
	if !floatsAlmostEqual(result.PRSScore, expected.PRSScore) {
		t.Errorf("PRSScore: got %v, want %v", result.PRSScore, expected.PRSScore)
	}
	if !reflect.DeepEqual(result.Details, expected.Details) {
		t.Errorf("Details: got %+v, want %+v", result.Details, expected.Details)
	}
}

// Invariance Integration Tests

func TestCalculatePRS_WithValidationEnabled_ValidInput(t *testing.T) {
	logging.SetSilentLoggingForTest()
	setupTestConfig(true, false) // Enable validation, disable strict mode

	// Valid SNPs
	snps := []model.AnnotatedSNP{
		{RSID: "rs1", Genotype: "AA", RiskAllele: "A", Beta: 0.2, Dosage: 2, Trait: "trait1"},
		{RSID: "rs2", Genotype: "AG", RiskAllele: "G", Beta: -0.3, Dosage: 1, Trait: "trait2"},
	}

	result, err := CalculatePRS(snps)
	if err != nil {
		t.Fatalf("CalculatePRS should succeed with valid input: %v", err)
	}

	expectedScore := 2*0.2 + 1*(-0.3) // 0.4 - 0.3 = 0.1
	if !floatsAlmostEqual(result.PRSScore, expectedScore) {
		t.Errorf("PRSScore: got %v, want %v", result.PRSScore, expectedScore)
	}
}

func TestCalculatePRS_WithValidationEnabled_InvalidDosage(t *testing.T) {
	logging.SetSilentLoggingForTest()
	setupTestConfig(true, false) // Enable validation, disable strict mode

	// Invalid dosage (>2)
	snps := []model.AnnotatedSNP{
		{RSID: "rs1", Genotype: "AA", RiskAllele: "A", Beta: 0.2, Dosage: 3, Trait: "trait1"}, // Invalid dosage
	}

	_, err := CalculatePRS(snps)
	if err == nil {
		t.Fatal("CalculatePRS should fail with invalid dosage")
	}

	if !strings.Contains(err.Error(), "Input validation failed") {
		t.Errorf("Expected 'Input validation failed' error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "rs1") {
		t.Errorf("Error should mention SNP rs1, got: %v", err)
	}
	if !strings.Contains(err.Error(), "pre-condition") {
		t.Errorf("Error should mention pre-condition phase, got: %v", err)
	}
}

func TestCalculatePRS_WithValidationEnabled_InvalidBeta(t *testing.T) {
	logging.SetSilentLoggingForTest()
	setupTestConfig(true, true) // Enable validation and strict mode

	// Invalid beta (NaN)
	snps := []model.AnnotatedSNP{
		{RSID: "rs1", Genotype: "AA", RiskAllele: "A", Beta: math.NaN(), Dosage: 2, Trait: "trait1"}, // NaN beta
	}

	_, err := CalculatePRS(snps)
	if err == nil {
		t.Fatal("CalculatePRS should fail with NaN beta")
	}

	if !strings.Contains(err.Error(), "Input validation failed") {
		t.Errorf("Expected 'Input validation failed' error, got: %v", err)
	}
}

func TestCalculatePRS_WithStrictMode_LargeBeta(t *testing.T) {
	logging.SetSilentLoggingForTest()
	setupTestConfig(true, true) // Enable validation and strict mode

	// Very large beta coefficient
	snps := []model.AnnotatedSNP{
		{RSID: "rs1", Genotype: "AA", RiskAllele: "A", Beta: 100.0, Dosage: 2, Trait: "trait1"}, // Very large beta
	}

	_, err := CalculatePRS(snps)
	if err == nil {
		t.Fatal("CalculatePRS should fail with extremely large beta in strict mode")
	}

	if !strings.Contains(err.Error(), "Input validation failed") {
		t.Errorf("Expected 'Input validation failed' error, got: %v", err)
	}
}

func TestCalculatePRS_ValidationDisabled_InvalidInput(t *testing.T) {
	logging.SetSilentLoggingForTest()
	setupTestConfig(false, false) // Disable validation

	// Invalid dosage should be allowed when validation is disabled
	snps := []model.AnnotatedSNP{
		{RSID: "rs1", Genotype: "AA", RiskAllele: "A", Beta: 0.2, Dosage: 3, Trait: "trait1"}, // Invalid dosage
	}

	result, err := CalculatePRS(snps)
	if err != nil {
		t.Fatalf("CalculatePRS should succeed when validation disabled: %v", err)
	}

	expectedScore := 3 * 0.2 // 0.6
	if !floatsAlmostEqual(result.PRSScore, expectedScore) {
		t.Errorf("PRSScore: got %v, want %v", result.PRSScore, expectedScore)
	}
}

func TestCalculatePRSWithBounds_ValidBounds(t *testing.T) {
	logging.SetSilentLoggingForTest()
	setupTestConfig(true, false) // Enable validation

	snps := []model.AnnotatedSNP{
		{RSID: "rs1", Genotype: "AA", RiskAllele: "A", Beta: 0.1, Dosage: 2, Trait: "trait1"},
		{RSID: "rs2", Genotype: "AG", RiskAllele: "G", Beta: -0.1, Dosage: 1, Trait: "trait2"},
	}

	result, err := CalculatePRSWithBounds(snps, -1.0, 1.0) // Bounds that include expected score
	if err != nil {
		t.Fatalf("CalculatePRSWithBounds should succeed with valid bounds: %v", err)
	}

	expectedScore := 2*0.1 + 1*(-0.1) // 0.2 - 0.1 = 0.1
	if !floatsAlmostEqual(result.PRSScore, expectedScore) {
		t.Errorf("PRSScore: got %v, want %v", result.PRSScore, expectedScore)
	}
}

func TestCalculatePRSWithBounds_InvalidBounds(t *testing.T) {
	logging.SetSilentLoggingForTest()
	setupTestConfig(true, false) // Enable validation

	snps := []model.AnnotatedSNP{
		{RSID: "rs1", Genotype: "AA", RiskAllele: "A", Beta: 0.5, Dosage: 2, Trait: "trait1"}, // Score will be 1.0
	}

	_, err := CalculatePRSWithBounds(snps, -0.5, 0.5) // Bounds that exclude expected score (1.0)
	if err == nil {
		t.Fatal("CalculatePRSWithBounds should fail when score exceeds bounds")
	}

	if !strings.Contains(err.Error(), "outside bounds") {
		t.Errorf("Expected 'outside bounds' error, got: %v", err)
	}
}

func TestValidationConfigFunctions(t *testing.T) {
	logging.SetSilentLoggingForTest()

	// Test enabling validation
	setupTestConfig(true, true)
	if !IsValidationEnabled() {
		t.Error("Validation should be enabled after setupTestConfig(true, true)")
	}
	if !IsStrictModeEnabled() {
		t.Error("Strict mode should be enabled after setupTestConfig(true, true)")
	}

	// Test disabling validation
	setupTestConfig(false, false)
	if IsValidationEnabled() {
		t.Error("Validation should be disabled after setupTestConfig(false, false)")
	}
	if IsStrictModeEnabled() {
		t.Error("Strict mode should be disabled after setupTestConfig(false, false)")
	}
}

func TestPRSCalculationError(t *testing.T) {
	snp := &model.AnnotatedSNP{RSID: "rs123", Beta: 0.1, Dosage: 2}

	// Test error with SNP
	err := &PRSCalculationError{
		Message: "test error",
		SNP:     snp,
		Phase:   "calculation",
		Cause:   nil,
	}

	errorStr := err.Error()
	if !strings.Contains(errorStr, "rs123") {
		t.Errorf("Error should contain SNP RSID, got: %s", errorStr)
	}
	if !strings.Contains(errorStr, "calculation") {
		t.Errorf("Error should contain phase, got: %s", errorStr)
	}
	if !strings.Contains(errorStr, "test error") {
		t.Errorf("Error should contain message, got: %s", errorStr)
	}

	// Test error without SNP
	err2 := &PRSCalculationError{
		Message: "general error",
		SNP:     nil,
		Phase:   "post-condition",
		Cause:   nil,
	}

	errorStr2 := err2.Error()
	if strings.Contains(errorStr2, "rs123") {
		t.Errorf("Error without SNP should not contain RSID, got: %s", errorStr2)
	}
	if !strings.Contains(errorStr2, "post-condition") {
		t.Errorf("Error should contain phase, got: %s", errorStr2)
	}
}
