package prs

import (
	"testing"
	"reflect"
	"phite.io/polygenic-risk-calculator/internal/model"
	"phite.io/polygenic-risk-calculator/internal/logging"
)


func floatsAlmostEqual(a, b float64) bool {
	const eps = 1e-9
	if a > b {
		return a-b < eps
	}
	return b-a < eps
}

func TestCalculatePRS_BasicSum(t *testing.T) {
	logging.SetSilentLoggingForTest()
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

	result := CalculatePRS(snps)

	if !floatsAlmostEqual(result.PRSScore, expected.PRSScore) {
		t.Errorf("PRSScore: got %v, want %v", result.PRSScore, expected.PRSScore)
	}
	if !reflect.DeepEqual(result.Details, expected.Details) {
		t.Errorf("Details: got %+v, want %+v", result.Details, expected.Details)
	}
}

func TestCalculatePRS_EmptyInput(t *testing.T) {
	logging.SetSilentLoggingForTest()
	result := CalculatePRS([]model.AnnotatedSNP{})
	if !floatsAlmostEqual(result.PRSScore, 0) {
		t.Errorf("Empty input: PRSScore = %v, want 0", result.PRSScore)
	}
	if len(result.Details) != 0 {
		t.Errorf("Empty input: Details length = %d, want 0", len(result.Details))
	}
}

func TestCalculatePRS_MissingSNPs(t *testing.T) {
	logging.SetSilentLoggingForTest()
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

	result := CalculatePRS(snps)
	if !floatsAlmostEqual(result.PRSScore, expected.PRSScore) {
		t.Errorf("PRSScore: got %v, want %v", result.PRSScore, expected.PRSScore)
	}
	if !reflect.DeepEqual(result.Details, expected.Details) {
		t.Errorf("Details: got %+v, want %+v", result.Details, expected.Details)
	}
}

func TestCalculatePRS_NegativeEffectSize(t *testing.T) {
	logging.SetSilentLoggingForTest()
	snps := []model.AnnotatedSNP{
		{RSID: "rsX", Genotype: "GG", RiskAllele: "G", Beta: -1.2, Dosage: 2, Trait: "traitX"},
	}

	expected := PRSResult{
		PRSScore: 2 * -1.2,
		Details: []SNPContribution{
			{"rsX", 2, -1.2, -2.4},
		},
	}

	result := CalculatePRS(snps)
	if !floatsAlmostEqual(result.PRSScore, expected.PRSScore) {
		t.Errorf("PRSScore: got %v, want %v", result.PRSScore, expected.PRSScore)
	}
	if !reflect.DeepEqual(result.Details, expected.Details) {
		t.Errorf("Details: got %+v, want %+v", result.Details, expected.Details)
	}
}
