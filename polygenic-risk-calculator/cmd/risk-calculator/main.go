package main

import (
	"fmt"
	"io"
	"os"

	"phite.io/polygenic-risk-calculator/internal/logging"

	"phite.io/polygenic-risk-calculator/internal/cli"
	"phite.io/polygenic-risk-calculator/internal/genotype"
	"phite.io/polygenic-risk-calculator/internal/gwas"
	"phite.io/polygenic-risk-calculator/internal/model"
	"phite.io/polygenic-risk-calculator/internal/output"
	"phite.io/polygenic-risk-calculator/internal/prs"
	"phite.io/polygenic-risk-calculator/internal/reference"
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

	opts, err := cli.ParseOptions(args)
	if err != nil {
		logging.Error("parameter error: %v", err)
		cli.PrintHelp()
		return 1
	}

	// Use canonical, validated SNP list from opts
	rsids := opts.SNPs

	gwasRecords, err := gwas.FetchGWASRecordsWithTable(opts.GWASDB, opts.GWASTable, rsids)
	if err != nil {
		logAndStderr(stderr, "Failed to load GWAS records: %v", err)
		return 1
	}
	logging.Info("Loaded %d GWAS records", len(gwasRecords))

	// Parse user genotype file
	logging.Info("Parsing genotype file: %s", opts.GenotypeFile)
	genoOut, err := genotype.ParseGenotypeData(genotype.ParseGenotypeDataInput{
		GenotypeFilePath: opts.GenotypeFile,
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
		AssociationsClean: gwas.MapToGWASList(gwasRecords),
	})
	logging.Info("Annotated %d SNPs", len(gwasOutput.AnnotatedSNPs))

	// Calculate PRS
	logging.Info("Calculating PRS")
	prsResult := prs.CalculatePRS(gwasOutput.AnnotatedSNPs)
	logging.Info("PRS calculation complete")

	// Load reference stats (optional)
	refDB := opts.ReferenceDB
	var refStats *model.ReferenceStats
	if refDB != "" {
		logging.Info("Loading reference stats from %s", refDB)
		refStatsRaw, _ := reference.LoadDefaultReferenceStats(refDB)
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
	logging.Info("Formatting output: format=%s, outputPath=%s", opts.Format, opts.Output)
	err = output.FormatOutput(norm, prsResult, summaries, genoOut.SNPsMissing, opts.Format, opts.Output, stdout)
	if err != nil {
		logging.Error("failed to format output: %v", err)
		return 1
	}
	logging.Info("Output formatting complete")

	return 0
}


func main() {
	logging.Info("PHITE CLI invoked")
	os.Exit(RunCLI(os.Args[1:], os.Stdout, os.Stderr))
}
