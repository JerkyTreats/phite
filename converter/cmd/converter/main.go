package main

import (
	"flag"
	"path/filepath"

	"github.com/JerkyTreats/PHITE/converter/internal/config"
	"github.com/JerkyTreats/PHITE/converter/internal/converter"
	"github.com/JerkyTreats/PHITE/converter/pkg/logger"
)

func main() {
	// Load configuration
	config, err := config.LoadConfig()
	if err != nil {
		logger.Fatal(err, "failed to load configuration")
	}

	inputFile := flag.String("input", "", "path to input TSV file")
	outputDir := flag.String("output-dir", config.GetOutputDir(), "directory to save JSON files")
	logLevel := flag.String("log-level", config.GetLogLevel(), "logging level (debug, info, error, fatal)")
	groupingMode := flag.String("grouping-mode", "group", "grouping mode: 'group' or 'topic'") // New flag
	flag.Parse()

	// Set logging level
	if err := logger.SetLevel(*logLevel); err != nil {
		logger.Fatal(err, "invalid log level specified")
	}

	if *inputFile == "" {
		logger.Fatal(nil, "input file is required")
	}

	// Create output directory if it doesn't exist
	absOutputDir, err := filepath.Abs(*outputDir)
	if err != nil {
		logger.Fatal(err, "failed to get absolute path for output directory")
	}

	// Save updated configuration if output directory was changed
	if *outputDir != config.GetOutputDir() {
		config.SetOutputDir(*outputDir)
		if err := config.Save(); err != nil {
			logger.Fatal(err, "failed to save configuration")
		}
	}

	parser := converter.NewTSVParser(*inputFile, absOutputDir, *groupingMode)
	outputFiles, errorRecords, err := parser.Parse()
	if err != nil {
		logger.Fatal(err, "failed to parse TSV")
	}

	if len(errorRecords) > 0 {
		logger.Info("some records were skipped due to invalid format", "errors", len(errorRecords))
	}

	if len(outputFiles) > 0 {
		logger.Info("conversion completed successfully", "input", *inputFile, "output-dir", absOutputDir)
		logger.Info("generated files:", "count", len(outputFiles))
		for _, file := range outputFiles {
			logger.Info("generated file", "path", file)
		}
	} else {
		logger.Fatal(nil, "no files were generated")
	}
}
