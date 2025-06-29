package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"phite.io/polygenic-risk-calculator/internal/config"
	snpsutil "phite.io/polygenic-risk-calculator/internal/snps"
)

type Options struct {
	GenotypeFile   string
	SNPs           []string
	SNPsFile       string
	GWASDB         string
	GWASTable      string
	Output         string
	Format         string
	ReferenceTable string
}

// ParseOptions parses CLI flags and resolves each parameter from CLI/env/config/default.
func ParseOptions(args []string) (Options, error) {
	flags := pflag.NewFlagSet("risk-calculator", pflag.ContinueOnError)

	var opts Options
	var snps string

	flags.StringVar(&opts.GenotypeFile, "genotype-file", "", "Path to genotype file (required)")
	flags.StringVar(&snps, "snps", "", "Comma-separated list of SNP IDs (required unless --snps-file)")
	flags.StringVar(&opts.SNPsFile, "snps-file", "", "Path to SNPs file (required unless --snps)")
	flags.StringVar(&opts.GWASDB, "gwas-db", "", "Path to GWAS DuckDB (required)")
	flags.StringVar(&opts.GWASTable, "gwas-table", "", "GWAS table name (optional)")
	flags.StringVar(&opts.Output, "output", "", "Output file path (optional)")
	flags.StringVar(&opts.Format, "format", "", "Output format (optional)")
	flags.StringVar(&opts.ReferenceTable, "reference-table", "reference_panel", "Reference stats table name (optional, default: reference_panel)")

	if err := flags.Parse(args); err != nil {
		return opts, err
	}

	// GWAS Database with Validation
	if opts.GWASDB != "" {
		config.Set("gwas_db_path", opts.GWASDB)
	}
	if opts.GWASTable != "" {
		config.Set("gwas_table", opts.GWASTable)
	}

	// Canonical SNP resolution (enforce mutual exclusion, requiredness, and validation)
	if opts.SNPsFile != "" {
		config.Set("snps_file", opts.SNPsFile)
	}
	if snps != "" {
		config.Set("snps", snps)
	}
	if snps != "" {
		opts.SNPs = strings.Split(snps, ",")
	}
	resolvedSNPs, err := snpsutil.ResolveSNPs(opts.SNPs, opts.SNPsFile)
	if err != nil {
		switch err {
		case snpsutil.ErrNoSNPsProvided:
			return opts, errors.New("either --snps or --snps-file is required")
		default:
			return opts, err
		}
	}
	opts.SNPs = resolvedSNPs
	if len(opts.SNPs) > 0 && opts.SNPsFile != "" && snps != "" {
		return opts, errors.New("--snps and --snps-file are mutually exclusive")
	}

	// Other validations
	errMsgs := []string{}
	if opts.GenotypeFile == "" {
		errMsgs = append(errMsgs, "--genotype-file or corresponding config key 'genotype_file' is required")
	}
	if opts.GWASDB == "" && config.GetString("gwas_db_path") == "" {
		errMsgs = append(errMsgs, "--gwas-db is required")
	}
	if len(errMsgs) > 0 {
		return opts, errors.New(strings.Join(errMsgs, "; "))
	}

	return opts, nil
}

// PrintHelp prints the usage/help text for the CLI.
func PrintHelp() {
	fmt.Fprintf(os.Stderr, `Usage: risk-calculator [OPTIONS]\n
Options:
  --genotype-file   Path to genotype file (required)
  --snps            Comma-separated list of SNP IDs (required unless --snps-file)
  --snps-file       Path to SNPs file (required unless --snps)
  --gwas-db         Path to GWAS DuckDB (required)
  --gwas-table      GWAS table name (optional)
  --output          Output file path (optional)
  --format          Output format (optional)
  --reference-db    Path to reference stats DB (optional)
`)
}
