package main

import (
	"io"
	"os"

	"phite.io/polygenic-risk-calculator/internal/cli"
	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/logging"
	"phite.io/polygenic-risk-calculator/internal/output"
	"phite.io/polygenic-risk-calculator/internal/pipeline"
	"phite.io/polygenic-risk-calculator/internal/prs"
)

// RunCLI parses arguments and runs the entrypoint logic. Returns exit code.
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

	pipelineInput := pipeline.PipelineInput{
		GenotypeFile:   opts.GenotypeFile,
		SNPs:           opts.SNPs,
		ReferenceTable: opts.ReferenceTable,
	}

	// Check for missing required keys early in RunCLI
	if len(config.MissingKeys) > 0 {
		logging.Error("Missing required configuration keys: %v", config.MissingKeys)
		return 1
	}

	outputData, err := pipeline.Run(pipelineInput)
	if err != nil {
		logging.Error("Pipeline error: %v", err)
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
