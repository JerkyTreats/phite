package output

import (
	"testing"
	"phite.io/polygenic-risk-calculator/internal/model"
	"phite.io/polygenic-risk-calculator/internal/prs"
)

func TestGenerateTraitSummaries(t *testing.T) {
	tests := []struct {
		name      string
		annotated []model.AnnotatedSNP
		norm      prs.NormalizedPRS
		want      []TraitSummary
	}{
		{
			name: "single trait, high risk",
			annotated: []model.AnnotatedSNP{
				{RSID: "rs1", Dosage: 2, Beta: 0.1, Trait: "BMI"},
				{RSID: "rs2", Dosage: 1, Beta: 0.2, Trait: "BMI"},
			},
			norm: prs.NormalizedPRS{RawScore: 0.4, ZScore: 2.0, Percentile: 95.0},
			want: []TraitSummary{{Trait: "BMI", NumRiskAlleles: 3, EffectWeightedContribution: 0.4, RiskLevel: "high"}},
		},
		{
			name: "multiple traits, moderate and low",
			annotated: []model.AnnotatedSNP{
				{RSID: "rs1", Dosage: 1, Beta: 0.1, Trait: "BMI"},
				{RSID: "rs2", Dosage: 2, Beta: 0.2, Trait: "Height"},
			},
			norm: prs.NormalizedPRS{RawScore: 0.5, ZScore: 0.0, Percentile: 50.0},
			want: []TraitSummary{{Trait: "BMI", NumRiskAlleles: 1, EffectWeightedContribution: 0.1, RiskLevel: "moderate"}, {Trait: "Height", NumRiskAlleles: 2, EffectWeightedContribution: 0.4, RiskLevel: "moderate"}},
		},
		{
			name: "missing trait info",
			annotated: []model.AnnotatedSNP{
				{RSID: "rs1", Dosage: 1, Beta: 0.1, Trait: ""},
			},
			norm: prs.NormalizedPRS{RawScore: 0.1, ZScore: -1.0, Percentile: 10.0},
			want: []TraitSummary{{Trait: "unknown", NumRiskAlleles: 1, EffectWeightedContribution: 0.1, RiskLevel: "low"}},
		},
		{
			name:      "empty input",
			annotated: nil,
			norm:      prs.NormalizedPRS{RawScore: 0, ZScore: 0, Percentile: 0},
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateTraitSummaries(tt.annotated, tt.norm)
			if len(got) != len(tt.want) {
				t.Errorf("got %d summaries, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got %+v, want %+v", got[i], tt.want[i])
				}
			}
		})
	}
}
