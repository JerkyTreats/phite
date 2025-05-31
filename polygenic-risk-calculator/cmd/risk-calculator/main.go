package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"phite.io/polygenic-risk-calculator/internal/model"
	"phite.io/polygenic-risk-calculator/internal/genotype"
	"phite.io/polygenic-risk-calculator/internal/gwas"
	"phite.io/polygenic-risk-calculator/internal/reference"
	"phite.io/polygenic-risk-calculator/internal/prs"
	"phite.io/polygenic-risk-calculator/internal/output"
)


// RunCLI parses arguments and runs the entrypoint logic. Returns exit code.
func RunCLI(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("polygenic-risk-calculator", flag.ContinueOnError)
	fs.SetOutput(stderr)
	genotypeFile := fs.String("genotype-file", "", "Path to user genotype file (required)")
	snps := fs.String("snps", "", "Comma-separated list of SNP rsids (required)")
	outputPath := fs.String("output", "", "Output file path (optional)")
	format := fs.String("format", "json", "Output format: json or csv (optional)")

	if err := fs.Parse(args); err != nil {
		return 2 // flag parse error
	}

	if *genotypeFile == "" || *snps == "" {
		fmt.Fprintln(stderr, "Error: --genotype-file and --snps are required.")
		fs.Usage()
		return 1
	}

	// Check if genotype file exists
	if _, err := os.Stat(*genotypeFile); err != nil {
		fmt.Fprintf(stderr, "Error: genotype file not found: %s\n", *genotypeFile)
		return 1
	}

		// Parse list of SNPs
	rsids := strings.Split(*snps, ",")
	for i := range rsids {
		rsids[i] = strings.TrimSpace(rsids[i])
	}
	if len(rsids) == 0 {
		fmt.Fprintln(stderr, "Error: no SNPs provided.")
		return 1
	}

	// Load GWAS associations from DuckDB
	gwasDB := os.Getenv("GWAS_DUCKDB")
	if gwasDB == "" {
		gwasDB = "internal/gwas/testdata/gwas.duckdb" // fallback for test/dev
	}
	gwasRecords, err := gwas.FetchGWASRecords(gwasDB, rsids)
	if err != nil {
		fmt.Fprintf(stderr, "Error: failed to load GWAS records: %v\n", err)
		return 1
	}

	// Parse user genotype file
	genoOut, err := genotype.ParseGenotypeData(genotype.ParseGenotypeDataInput{
		GenotypeFilePath: *genotypeFile,
		RequestedRSIDs:   rsids,
		GWASData:         gwasRecords,
	})
	if err != nil {
		fmt.Fprintf(stderr, "Error: failed to parse genotype file: %v\n", err)
		return 1
	}

	// Annotate SNPs
	gwasOutput := gwas.FetchAndAnnotateGWAS(gwas.GWASDataFetcherInput{
		ValidatedSNPs:     genoOut.ValidatedSNPs,
		AssociationsClean: mapToGWASList(gwasRecords),
	})

	// Calculate PRS
	prsResult := prs.CalculatePRS(gwasOutput.AnnotatedSNPs)

	// Load reference stats (optional)
	refDB := os.Getenv("REFERENCE_DUCKDB")
	var refStats *model.ReferenceStats
	if refDB != "" {
		// For demo: use EUR/height/v1 as default; in real CLI, expose as flags
		refStatsRaw, _ := reference.LoadReferenceStatsFromDuckDB(refDB, "EUR", "height", "v1")
		if refStatsRaw != nil {
			refStats = &model.ReferenceStats{
				Mean:     refStatsRaw.Mean,
				Std:      refStatsRaw.Std,
				Min:      refStatsRaw.Min,
				Max:      refStatsRaw.Max,
				Ancestry: refStatsRaw.Ancestry,
				Trait:    refStatsRaw.Trait,
				Model:    refStatsRaw.Model,
			}
		}
	}

	var norm prs.NormalizedPRS
	if refStats != nil {
		norm, _ = prs.NormalizePRS(prsResult, *refStats)
	}

	// Generate trait summaries
	summaries := output.GenerateTraitSummaries(gwasOutput.AnnotatedSNPs, norm)

	// Output results
	err = output.FormatOutput(norm, prsResult, summaries, genoOut.SNPsMissing, *format, *outputPath, stdout)
	if err != nil {
		fmt.Fprintf(stderr, "Error: failed to format output: %v\n", err)
		return 1
	}

	return 0
}

// mapToGWASList converts a map of GWASSNPRecord to a slice for annotation
func mapToGWASList(m map[string]model.GWASSNPRecord) []model.GWASSNPRecord {
	if m == nil {
		return nil
	}
	records := make([]model.GWASSNPRecord, 0, len(m))
	for _, rec := range m {
		records = append(records, rec)
	}
	return records
}




func main() {
	os.Exit(RunCLI(os.Args[1:], os.Stdout, os.Stderr))
}
