package output

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"phite.io/polygenic-risk-calculator/internal/logging"
	"io"
	"os"

	"phite.io/polygenic-risk-calculator/internal/prs"
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
	logging.Info("Formatting output: format=%s, outFile=%s", format, outFile)
	if format != "json" && format != "csv" {
		logging.Error("unsupported output format: %s", format)
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
			logging.Error("failed to create output file: %v", err)
			return err
		}
		defer f.Close()
		logging.Info("Writing output to file: %s", outFile)
		w = f
	} else if out != nil {
		w = out
	} else {
		w = os.Stdout
	}

	if format == "json" {
		logging.Info("Encoding output as JSON")
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		if err := e.Encode(output); err != nil {
			logging.Error("failed to encode output as JSON: %v", err)
			return err
		}
		return nil
	}

	// CSV: only prs.NormalizedPRS and prs.PRSResult are written as flat tables; others as JSON-encoded fields
	logging.Info("Encoding output as CSV")
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
		b, err := json.Marshal(summaries)
		if err != nil {
			logging.Error("failed to marshal trait summaries as JSON: %v", err)
		}
		csvw.Write([]string{"trait_summaries", string(b)})
	}
	// Write SNPSMissing as JSON
	if snpsMissing != nil {
		b, err := json.Marshal(snpsMissing)
		if err != nil {
			logging.Error("failed to marshal snps_missing as JSON: %v", err)
		}
		csvw.Write([]string{"snps_missing", string(b)})
	}
	return nil
}
