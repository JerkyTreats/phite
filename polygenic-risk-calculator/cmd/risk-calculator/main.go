package main

import (
	"flag"
	"fmt"
	"phite.io/polygenic-risk-calculator/internal/logging"
	"io"
	"os"
	"strings"

	"phite.io/polygenic-risk-calculator/internal/genotype"
	"phite.io/polygenic-risk-calculator/internal/gwas"
	"phite.io/polygenic-risk-calculator/internal/model"
	"phite.io/polygenic-risk-calculator/internal/output"
	"phite.io/polygenic-risk-calculator/internal/prs"
	"phite.io/polygenic-risk-calculator/internal/reference"
	"phite.io/polygenic-risk-calculator/internal/snps"
)

// RunCLI parses arguments and runs the entrypoint logic. Returns exit code.
func logAndStderr(stderr io.Writer, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logging.Error("%s", msg)
	fmt.Fprintln(stderr, msg)
}

func RunCLI(args []string, stdout, stderr io.Writer) int {
	logging.Info("PHITE CLI started with args: %v", args)
	defer func() {
		logging.Info("PHITE CLI exiting")
	}()
	fs := flag.NewFlagSet("polygenic-risk-calculator", flag.ContinueOnError)
	fs.SetOutput(stderr)
	genotypeFile := fs.String("genotype-file", "", "Path to user genotype file (required)")
	snpsFlag := fs.String("snps", "", "Comma-separated list of SNP rsids (required unless --snps-file is set)")
	snpsFileFlag := fs.String("snps-file", "", "Path to file containing SNP rsids (JSON or CSV, mutually exclusive with --snps)")
	outputPath := fs.String("output", "", "Output file path (optional)")
	format := fs.String("format", "json", "Output format: json or csv (optional)")

	logging.Info("Parsing CLI flags")
	if err := fs.Parse(args); err != nil {
		logging.Error("flag parse error: %v", err)
		return 2 // flag parse error
	}

	// Enforce mutual exclusivity and required flags
	if *snpsFlag != "" && *snpsFileFlag != "" {
		logging.Error("--snps and --snps-file are mutually exclusive. Provide only one.")
		fs.Usage()
		return 1
	}
	if *snpsFlag == "" && *snpsFileFlag == "" {
		logging.Error("one of --snps or --snps-file is required.")
		fs.Usage()
		return 1
	}
	if *genotypeFile == "" {
		logging.Error("--genotype-file is required.")
		fs.Usage()
		return 1
	}

	// Check if genotype file exists
	if _, err := os.Stat(*genotypeFile); err != nil {
		logAndStderr(stderr, "genotype file not found: %s", *genotypeFile)
		return 1
	}

	// Parse list of SNPs
	var rsids []string
	if *snpsFlag != "" {
		rsids = strings.Split(*snpsFlag, ",")
		for i := range rsids {
			rsids[i] = strings.TrimSpace(rsids[i])
		}
		// Deduplicate and validate
		seen := make(map[string]struct{})
		out := make([]string, 0, len(rsids))
		for _, r := range rsids {
			if r == "" {
				logging.Error("empty rsid in --snps list.")
				return 1
			}
			if _, exists := seen[r]; !exists {
				seen[r] = struct{}{}
				out = append(out, r)
			}
		}
		rsids = out
	} else {
		var err error
		rsids, err = snps.ParseSNPsFromFile(*snpsFileFlag)
		if err != nil {
			logging.Error("failed to parse SNPs from file: %v", err)
			return 1
		}
		logging.Info("Parsed %d SNP rsids from file %s", len(rsids), *snpsFileFlag)
	}
	if len(rsids) == 0 {
		logging.Error("no SNPs provided.")
		return 1
	}

	// Load GWAS associations from DuckDB
	gwasDB := os.Getenv("GWAS_DUCKDB")
	if gwasDB == "" {
		gwasDB = "internal/gwas/testdata/gwas.duckdb" // fallback for test/dev
	}
	logging.Info("Loading GWAS records from %s", gwasDB)
	gwasRecords, err := gwas.FetchGWASRecords(gwasDB, rsids)
	if err != nil {
		logging.Error("failed to load GWAS records: %v", err)
		return 1
	}
	logging.Info("Loaded %d GWAS records", len(gwasRecords))

	// Parse user genotype file
	logging.Info("Parsing genotype file: %s", *genotypeFile)
	genoOut, err := genotype.ParseGenotypeData(genotype.ParseGenotypeDataInput{
		GenotypeFilePath: *genotypeFile,
		RequestedRSIDs:   rsids,
		GWASData:         gwasRecords,
	})
	if err != nil {
		logging.Error("failed to parse genotype file: %v", err)
		return 1
	}
	logging.Info("Parsed genotype file, validated %d SNPs", len(genoOut.ValidatedSNPs))

	// Annotate SNPs
	logging.Info("Annotating SNPs with GWAS associations")
	gwasOutput := gwas.FetchAndAnnotateGWAS(gwas.GWASDataFetcherInput{
		ValidatedSNPs:     genoOut.ValidatedSNPs,
		AssociationsClean: mapToGWASList(gwasRecords),
	})
	logging.Info("Annotated %d SNPs", len(gwasOutput.AnnotatedSNPs))

	// Calculate PRS
	logging.Info("Calculating PRS")
	prsResult := prs.CalculatePRS(gwasOutput.AnnotatedSNPs)
	logging.Info("PRS calculation complete")

	// Load reference stats (optional)
	refDB := os.Getenv("REFERENCE_DUCKDB")
	var refStats *model.ReferenceStats
	if refDB != "" {
		logging.Info("Loading reference stats from %s", refDB)
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
			logging.Info("Loaded reference stats for ancestry=%s, trait=%s, model=%s", refStats.Ancestry, refStats.Trait, refStats.Model)
		} else {
			logging.Info("No reference stats found in %s", refDB)
		}
	}

	var norm prs.NormalizedPRS
	if refStats != nil {
		logging.Info("Normalizing PRS using reference stats")
		norm, _ = prs.NormalizePRS(prsResult, *refStats)
		logging.Info("PRS normalization complete")
	}

	// Generate trait summaries
	logging.Info("Generating trait summaries")
	summaries := output.GenerateTraitSummaries(gwasOutput.AnnotatedSNPs, norm)
	logging.Info("Generated %d trait summaries", len(summaries))

	// Output results
	logging.Info("Formatting output: format=%s, outputPath=%s", *format, *outputPath)
	err = output.FormatOutput(norm, prsResult, summaries, genoOut.SNPsMissing, *format, *outputPath, stdout)
	if err != nil {
		logging.Error("failed to format output: %v", err)
		return 1
	}
	logging.Info("Output formatting complete")

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
	logging.Info("PHITE CLI invoked")
	os.Exit(RunCLI(os.Args[1:], os.Stdout, os.Stderr))
}
