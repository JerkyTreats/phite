package output

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"phite.io/polygenic-risk-calculator/prs"
)

// TraitSummary represents a summary for a trait (from data_model.md)
type TraitSummary struct {
	Trait                      string  `json:"trait"`
	NumRiskAlleles             int     `json:"num_risk_alleles"`
	EffectWeightedContribution float64 `json:"effect_weighted_contribution"`
	RiskLevel                  string  `json:"risk_level"`
}

// OutputResult represents the output structure for PRS results, summaries, and missing SNPs.
type OutputResult struct {
	NormalizedPRS  prs.NormalizedPRS `json:"normalized_prs"`
	PRSResult      prs.PRSResult     `json:"prs_result"`
	TraitSummaries []TraitSummary    `json:"trait_summaries"`
	SNPSMissing    []string          `json:"snps_missing"`
}

// FormatOutput serializes results as JSON or CSV and writes to file or stdout.
// If outFile is empty, writes to out (or stdout if out is nil).
func FormatOutput(norm prs.NormalizedPRS, prs prs.PRSResult, summaries []TraitSummary, snpsMissing []string, format, outFile string, out io.Writer) error {
	if format != "json" && format != "csv" {
		return errors.New("unsupported format: must be 'json' or 'csv'")
	}

	output := OutputResult{
		NormalizedPRS:  norm,
		PRSResult:      prs,
		TraitSummaries: summaries,
		SNPSMissing:    snpsMissing,
	}

	var w io.Writer
	if outFile != "" {
		f, err := os.Create(outFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		w = f
	} else if out != nil {
		w = out
	} else {
		w = os.Stdout
	}

	if format == "json" {
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		return e.Encode(output)
	}

	// CSV: only prs.NormalizedPRS and prs.PRSResult are written as flat tables; others as JSON-encoded fields
	csvw := csv.NewWriter(w)
	defer csvw.Flush()
	// Write prs.NormalizedPRS
	csvw.Write([]string{"raw_score", "z_score", "percentile"})
	csvw.Write([]string{
		fmt.Sprintf("%v", norm.RawScore),
		fmt.Sprintf("%v", norm.ZScore),
		fmt.Sprintf("%v", norm.Percentile),
	})
	// Write prs.PRSResult
	csvw.Write([]string{"prs_score"})
	csvw.Write([]string{fmt.Sprintf("%v", prs.PRSScore)})
	// Write TraitSummaries as JSON
	if summaries != nil {
		b, _ := json.Marshal(summaries)
		csvw.Write([]string{"trait_summaries", string(b)})
	}
	// Write SNPSMissing as JSON
	if snpsMissing != nil {
		b, _ := json.Marshal(snpsMissing)
		csvw.Write([]string{"snps_missing", string(b)})
	}
	return nil
}
