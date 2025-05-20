// Package converter provides functionality for parsing TSV files containing genetic
// SNP data and converting them into structured JSON format. The parser handles
// various validation scenarios and provides detailed error messages for invalid
// input formats.
package converter

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JerkyTreats/PHITE/converter/internal/config"
	"github.com/JerkyTreats/PHITE/converter/internal/models"
	"github.com/JerkyTreats/PHITE/converter/pkg/logger"
)

// ParseResult contains both valid records and any error records encountered during parsing
type ParseResult struct {
	Result       *models.ConversionResult
	ErrorRecords []string
}

// TSVParser is a parser for TSV files containing genetic SNP data.
// It converts the TSV format into a structured JSON format with groupings of SNPs.
type TSVParser struct {
	inputFile string
	outputDir string
	config    config.Config
}

// SaveResult saves a ConversionResult to the specified output file in JSON format.
//
// Args:
//
//	result: The ConversionResult to save
//	outputFile: Path to the output file
//
// Returns:
//
//	error: If any error occurs during saving
//
// Possible errors:
//   - os.ErrPermission: If insufficient permissions to write to file
//   - json.MarshalIndent: If JSON encoding fails
//   - os.WriteFile: If file write operation fails
func SaveResult(result *models.ConversionResult, outputFile string) error {
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		logger.Error(err, "failed to marshal JSON")
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	err = os.WriteFile(outputFile, jsonBytes, 0644)
	if err != nil {
		logger.Error(err, "failed to write output file")
		return fmt.Errorf("failed to write output file: %w", err)
	}

	logger.Info("saved result to file", "file", outputFile)
	return nil
}

// NewTSVParser creates a new TSVParser instance with the specified input file.
// The input file should be a TSV file with the following columns:
// Topic, Group, Gene, RS ID, Allele, Subject Genotype, Notes
func NewTSVParser(inputFile string, outputDir string) *TSVParser {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal(err, "failed to load configuration")
	}
	return &TSVParser{
		inputFile: inputFile,
		outputDir: outputDir,
		config:    cfg,
	}
}

// Parse reads and parses the TSV file into a structured JSON format.
// The input TSV should have the following columns:
// Topic, Group, Gene, RS ID, Allele, Subject Genotype, Notes
//
// The parser performs the following operations:
// 1. Validates the input file exists and can be read
// 2. Reads all records from the TSV file
// 3. Skips the header row
// 4. Groups SNPs by their Group field
// 5. For each SNP:
//   - Validates record format (must have 7 columns)
//   - Skips SNPs with blank or "--" genotypes
//   - Determines genotype match (None/Partial/Full)
//
// 6. Returns a ConversionResult containing all valid SNPs grouped by their Group
//
// Returns:
// - *models.ConversionResult: The parsed and structured SNP data
// - error: If any errors occur during parsing
//
// Possible errors:
// - os.ErrNotExist: If the input file does not exist
// - io.EOF: If the file is empty
// - csv.ParseError: If the file cannot be parsed as TSV
// - fmt.Errorf: For invalid record formats or other parsing errors
func (p *TSVParser) Parse() ([]string, []string, error) {
	file, err := os.Open(p.inputFile)
	if err != nil {
		logger.Error(err, "failed to open file")
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create map to store filenames for each grouping
	groupFilenames := make(map[string]string)

	reader := csv.NewReader(file)
	reader.Comma = '\t' // TSV format

	records, err := reader.ReadAll()
	if err != nil {
		logger.Error(err, "failed to read TSV")
		return nil, nil, fmt.Errorf("failed to read TSV: %w", err)
	}

	// Skip header
	if len(records) == 0 {
		logger.Error(nil, "empty file")
		return nil, nil, fmt.Errorf("empty file")
	}
	records = records[1:]

	// Create map to group SNPs by Group
	groupings := make(map[string]*models.Grouping)

	var errorRecords []string
	var outputFiles []string

	// Ensure output directory exists
	if err := os.MkdirAll(p.outputDir, 0755); err != nil {
		logger.Error(err, "failed to create output directory")
		return nil, nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, record := range records {
		logger.Debug("Processing record", "record", record)

		if len(record) != 7 {
			errorRecords = append(errorRecords, fmt.Sprintf("Record with %d columns: %v", len(record), record))
			logger.Info("Invalid record format", "record", record)
			continue
		}

		group := record[1]    // Group column
		topic := record[0]    // Topic column
		genotype := record[5] // Genotype column

		
		// Create new grouping if it doesn't exist
		if _, exists := groupings[group]; !exists {
			newGroup := &models.Grouping{
				Topic: topic,
				Name:  group,
			}
			groupings[group] = newGroup
			logger.Debug("New group created", "group", newGroup)
		}

		// Create and validate SNP
		snp, err := models.NewSNP(
			record[2], // Gene
			record[3], // RSID
			record[4], // Allele
			record[6], // Notes
			genotype,  // Genotype
		)

		if err != nil {
			errorRecords = append(errorRecords, fmt.Sprintf("Record validation failed: %v", record))
			logger.Info("Skipping SNP due to validation error", "error", err)
			continue
		}

		logger.Debug("SNP created", "snp", snp)
		groupings[group].SNP = append(groupings[group].SNP, *snp)
		logger.Debug("SNP added to group", "group", group, "snp", snp)
	}

	// Save each grouping to its own JSON file
	for _, grouping := range groupings {
		// Generate filename-safe version of group name
		filename := fmt.Sprintf("%s.json", strings.ReplaceAll(grouping.Name, "/", "-"))
		outputPath := filepath.Join(p.outputDir, filename)

		// Use AddIfMatch to build the output SNP slice
		var filteredSNPs []models.SNP
		for _, snp := range grouping.SNP {
			filteredSNPs = models.AddIfMatch(filteredSNPs, snp, p.config.GetMatchLevel())
		}

		if err := SaveResult(&models.ConversionResult{Grouping: models.Grouping{
			Topic: grouping.Topic,
			Name:  grouping.Name,
			SNP:   filteredSNPs,
		}}, outputPath); err != nil {
			logger.Error(err, "failed to save grouping", "group", grouping.Name)
			return nil, nil, fmt.Errorf("failed to save grouping %s: %w", grouping.Name, err)
		}
		outputFiles = append(outputFiles, outputPath)
		groupFilenames[grouping.Name] = filename
	}

	// Log the filename mappings
	for groupName, filename := range groupFilenames {
		if groupName != filename[:len(filename)-5] { // Remove .json extension
			logger.Info("group name contains special characters", "group", groupName, "file", filename)
		}
	}

	if len(errorRecords) > 0 {
		logger.Info("some records were skipped due to invalid format", "errors", len(errorRecords))
	}

	logger.Info("parsing completed", "files", len(outputFiles), "errors", len(errorRecords))
	return outputFiles, errorRecords, nil
}
