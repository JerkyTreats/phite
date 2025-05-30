package prs

import (
	"testing"
	"reflect"
)


func floatsAlmostEqual(a, b float64) bool {
	const eps = 1e-9
	if a > b {
		return a-b < eps
	}
	return b-a < eps
}

func TestCalculatePRS_BasicSum(t *testing.T) {
	snps := []AnnotatedSNP{
		{"rs1", "AA", "A", 0.2, 2, "trait1"},
		{"rs2", "AG", "G", -0.5, 1, "trait2"},
		{"rs3", "TT", "T", 0.1, 0, "trait3"},
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
	result := CalculatePRS([]AnnotatedSNP{})
	if !floatsAlmostEqual(result.PRSScore, 0) {
		t.Errorf("Empty input: PRSScore = %v, want 0", result.PRSScore)
	}
	if len(result.Details) != 0 {
		t.Errorf("Empty input: Details length = %d, want 0", len(result.Details))
	}
}

func TestCalculatePRS_MissingSNPs(t *testing.T) {
	// Simulate missing SNP by omitting from input; should just not contribute
	snps := []AnnotatedSNP{
		{"rs1", "AA", "A", 0.3, 2, "trait1"},
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
	snps := []AnnotatedSNP{
		{"rsX", "GG", "G", -1.2, 2, "traitX"},
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
