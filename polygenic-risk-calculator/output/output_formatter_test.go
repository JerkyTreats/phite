package output

import (
	"os"
	"strings"
	"testing"
	"phite.io/polygenic-risk-calculator/prs"
)

// prs.NormalizedPRS and PRSResult are imported from prs package.


func TestOutputFormatter_JSON_ToStdout(t *testing.T) {
	norm := prs.NormalizedPRS{RawScore: 1.1, ZScore: 0.5, Percentile: 70.0}
	prs := prs.PRSResult{PRSScore: 1.1, Details: nil}
	summaries := []TraitSummary{{Trait: "height", NumRiskAlleles: 5, EffectWeightedContribution: 0.8, RiskLevel: "moderate"}}
	snps := []string{"rs1", "rs2"}
	var out strings.Builder
	err := FormatOutput(norm, prs, summaries, snps, "json", "", &out)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "\"raw_score\"") {
		t.Errorf("JSON output missing expected field: %v", out.String())
	}
}

func TestOutputFormatter_CSV_ToStdout(t *testing.T) {
	norm := prs.NormalizedPRS{RawScore: 2.2, ZScore: 1.5, Percentile: 95.0}
	prs := prs.PRSResult{PRSScore: 2.2, Details: nil}
	summaries := []TraitSummary{{Trait: "BMI", NumRiskAlleles: 3, EffectWeightedContribution: 0.6, RiskLevel: "high"}}
	snps := []string{"rs3"}
	var out strings.Builder
	err := FormatOutput(norm, prs, summaries, snps, "csv", "", &out)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "raw_score,z_score,percentile") {
		t.Errorf("CSV output missing header: %v", out.String())
	}
}

func TestOutputFormatter_ToFile(t *testing.T) {
	norm := prs.NormalizedPRS{RawScore: 3.3, ZScore: 2.5, Percentile: 99.0}
	prs := prs.PRSResult{PRSScore: 3.3, Details: nil}
	summaries := []TraitSummary{}
	snps := []string{}
	file := "test_output.json"
	defer os.Remove(file)
	err := FormatOutput(norm, prs, summaries, snps, "json", file, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	b, err := os.ReadFile(file)
	if err != nil {
		t.Errorf("failed to read output file: %v", err)
	}
	if !strings.Contains(string(b), "raw_score") {
		t.Errorf("Output file missing expected field: %v", string(b))
	}
}

func TestOutputFormatter_HandlesEmptyInputs(t *testing.T) {
	err := FormatOutput(prs.NormalizedPRS{}, prs.PRSResult{}, nil, nil, "json", "", nil)
	if err != nil {
		t.Errorf("unexpected error for empty input: %v", err)
	}
}
