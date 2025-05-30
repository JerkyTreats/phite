package main

import (
	"flag"
	"fmt"
	"io"
	"os"
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

	// TODO: Implement pipeline orchestration and error handling as per brief
	_ = *outputPath
	_ = *format
	fmt.Fprintln(stdout, "Entrypoint stub: arguments parsed successfully.")
	return 0
}

func main() {
	os.Exit(RunCLI(os.Args[1:], os.Stdout, os.Stderr))
}
