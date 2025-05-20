package main

import (
	"flag"

	"github.com/JerkyTreats/PHITE/converter/internal/converter"
	"github.com/JerkyTreats/PHITE/converter/pkg/logger"
)

func main() {
	inputFile := flag.String("input", "", "path to input TSV file")
	outputFile := flag.String("output", "output.json", "path to output JSON file")
	logLevel := flag.String("log-level", "info", "logging level (debug, info, error, fatal)")
	flag.Parse()

	// Set logging level
	if err := logger.SetLevel(*logLevel); err != nil {
		logger.Fatal(err, "invalid log level specified")
	}

	if *inputFile == "" {
		logger.Fatal(nil, "input file is required")
	}

	parser := converter.NewTSVParser(*inputFile)
	result, err := parser.Parse()
	if err != nil {
		logger.Fatal(err, "failed to parse TSV")
	}

	err = converter.SaveResult(result, *outputFile)
	if err != nil {
		logger.Fatal(err, "failed to save result")
	}

	logger.Info("conversion completed successfully", "input", *inputFile, "output", *outputFile)
}
