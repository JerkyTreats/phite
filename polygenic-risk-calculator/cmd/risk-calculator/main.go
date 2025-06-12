package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"phite.io/polygenic-risk-calculator/internal/cli"
	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/db"
	"phite.io/polygenic-risk-calculator/internal/logging"
	"phite.io/polygenic-risk-calculator/internal/output"
	"phite.io/polygenic-risk-calculator/internal/pipeline"
	"phite.io/polygenic-risk-calculator/internal/prs"
)

// RunCLI parses arguments and runs the entrypoint logic. Returns exit code.
func logAndStderr(stderr io.Writer, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logging.Error("%s", msg)
	fmt.Fprintln(stderr, msg)
}

func RunCLI(args []string, stdout, stderr io.Writer) int {
	logging.Info("PHITE CLI started with args: %v", args)

	// Validate configuration early
	if err := config.Validate(); err != nil {
		logAndStderr(stderr, "Configuration error: %v", err)
		return 1
	}
	defer func() {
		logging.Info("PHITE CLI exiting")
	}()

	// Initial repository creation
	dbType := "duckdb"
	dbPath := "gwas/gwas.duckdb"
	repo, err := db.GetRepository(context.Background(), dbType, map[string]string{
		"path": dbPath,
	})
	if err != nil {
		logAndStderr(stderr, "DB error: %v", err)
		return 1
	}

	opts, err := cli.ParseOptions(args, repo)
	if err != nil {
		logging.Error("parameter error: %v", err)
		cli.PrintHelp()
		return 1
	}

	// Later re-initialized with actual path from options
	if opts.GWASDB != "" {
		repo, err = db.GetRepository(context.Background(), dbType, map[string]string{
			"path": opts.GWASDB,
		})
		if err != nil {
			logAndStderr(stderr, "DB error: %v", err)
			return 1
		}
		opts.Repo = repo
	}

	// Use canonical, validated SNP list from opts
	// Handle missing Ancestry/Model fields for backward compatibility
	ancestry := ""
	model := ""
	// cli.Options does not define Ancestry or Model; pass empty string for now

	pipelineInput := pipeline.PipelineInput{
		GenotypeFile:   opts.GenotypeFile,
		SNPs:           opts.SNPs,
		GWASTable:      opts.GWASTable,
		ReferenceTable: opts.ReferenceTable,
		Ancestry:       ancestry,
		Model:          model,
	}

	outputData, err := pipeline.Run(pipelineInput)
	if err != nil {
		logAndStderr(stderr, "Pipeline error: %v", err)
		return 1
	}

	// Output results (formatting)
	logging.Info("Formatting output: format=%s, outputPath=%s", opts.Format, opts.Output)
	// For compatibility, output only the first trait's PRS and normalized PRS if present
	var normPRSVal interface{} = nil
	var prsResultVal interface{} = nil
	if len(outputData.NormalizedPRS) > 0 {
		for _, v := range outputData.NormalizedPRS {
			normPRSVal = v
			break
		}
	}
	if len(outputData.PRSResults) > 0 {
		for _, v := range outputData.PRSResults {
			prsResultVal = v
			break
		}
	}
	var normPRS prs.NormalizedPRS
	var prsResult prs.PRSResult
	if normPRSVal != nil {
		normPRS = normPRSVal.(prs.NormalizedPRS)
	}
	if prsResultVal != nil {
		prsResult = prsResultVal.(prs.PRSResult)
	}
	err = output.FormatOutput(
		normPRS,
		prsResult,
		outputData.TraitSummaries,
		outputData.SNPSMissing,
		opts.Format,
		opts.Output,
		stdout,
	)
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
