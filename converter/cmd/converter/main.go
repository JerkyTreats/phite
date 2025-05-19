package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/JerkyTreats/PHITE/converter/internal/converter"
	"github.com/JerkyTreats/PHITE/converter/pkg/logger"
)

func main() {
	inputFile := flag.String("input", "", "path to input TSV file")
	outputFile := flag.String("output", "output.json", "path to output JSON file")
	flag.Parse()

	if *inputFile == "" {
		logger.Fatal(nil, "input file is required")
	}

	parser := converter.NewTSVParser(*inputFile)
	result, err := parser.Parse()
	if err != nil {
		logger.Fatal(err, "failed to parse TSV")
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		logger.Fatal(err, "failed to marshal JSON")
	}

	err = os.WriteFile(*outputFile, jsonBytes, 0644)
	if err != nil {
		logger.Fatal(err, "failed to write output file")
	}

	logger.Info("conversion completed successfully", "input", *inputFile, "output", *outputFile)
}
