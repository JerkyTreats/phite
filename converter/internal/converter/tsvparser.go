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
	inputFile    string
	outputDir    string
	config       config.Config
	groupingMode string // "group" or "topic"
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
func saveTopicOutput(topicOutput *models.TopicOutput, outputFile string) error {
	jsonBytes, err := json.MarshalIndent(topicOutput, "", "  ")
	if err != nil {
		logger.Error(err, "failed to marshal TopicOutput JSON")
		return fmt.Errorf("failed to marshal TopicOutput JSON: %w", err)
	}

	err = os.WriteFile(outputFile, jsonBytes, 0644)
	if err != nil {
		logger.Error(err, "failed to write TopicOutput file")
		return fmt.Errorf("failed to write TopicOutput file: %w", err)
	}

	logger.Info("saved TopicOutput to file", "file", outputFile)
	return nil
}

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
func NewTSVParser(inputFile string, outputDir string, groupingMode string) *TSVParser {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal(err, "failed to load configuration")
	}
	return &TSVParser{
		inputFile:    inputFile,
		outputDir:    outputDir,
		config:       cfg,
		groupingMode: groupingMode,
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
	// Validate groupingMode
	if p.groupingMode != "group" && p.groupingMode != "topic" {
		errMsg := fmt.Sprintf("invalid grouping mode: %s. Must be 'group' or 'topic'", p.groupingMode)
		logger.Error(nil, errMsg)
		return nil, nil, fmt.Errorf("invalid grouping mode: %s. Must be 'group' or 'topic'", p.groupingMode)
	}

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

	var errorRecords []string
	var outputFiles []string

	// Ensure output directory exists
	if err := os.MkdirAll(p.outputDir, 0755); err != nil {
		logger.Error(err, "failed to create output directory")
		return nil, nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	if p.groupingMode == "topic" {
		// Data structure for topic mode: map[topicName] -> models.TopicOutput
		topicsData := make(map[string]*models.TopicOutput)

		for _, record := range records {
			logger.Debug("Processing record for topic mode", "record", record)
			if len(record) != 7 {
				errorRecords = append(errorRecords, fmt.Sprintf("Record with %d columns: %v", len(record), record))
				logger.Info("Invalid record format", "record", record)
				continue
			}

			topicName := record[0]
			groupName := record[1]
			genotype := record[5]

			// Create new TopicOutput if it doesn't exist for this topicName
			if _, exists := topicsData[topicName]; !exists {
				topicsData[topicName] = &models.TopicOutput{
					Topic:     topicName,
					Groupings: make(map[string][]models.SNP),
				}
				logger.Debug("New topic entry created", "topicName", topicName)
			}

			// Ensure the specific group map exists within the topic's groupings
			if _, exists := topicsData[topicName].Groupings[groupName]; !exists {
				topicsData[topicName].Groupings[groupName] = []models.SNP{}
			}

			snp, err := models.NewSNP(record[2], record[3], record[4], record[6], genotype)
			if err != nil {
				errorRecords = append(errorRecords, fmt.Sprintf("Record validation failed: %v", record))
				logger.Info("Skipping SNP due to validation error", "error", err)
				continue
			}

			topicsData[topicName].Groupings[groupName] = append(topicsData[topicName].Groupings[groupName], *snp)
			logger.Debug("SNP added to topic-group", "topicName", topicName, "groupName", groupName, "snpRSID", snp.RSID)
		}

		// Save each topic's data to its own JSON file
		for topicName, topicOutputData := range topicsData {
			filename := fmt.Sprintf("%s.json", strings.ReplaceAll(topicName, "/", "-"))
			outPath := filepath.Join(p.outputDir, filename)

			// Apply AddIfMatch filtering to each group's SNPs within the topic
			filteredGroupings := make(map[string][]models.SNP)
			for groupName, snpList := range topicOutputData.Groupings {
				var filteredSNPs []models.SNP
				for _, snp := range snpList {
					filteredSNPs = models.AddIfMatch(filteredSNPs, snp, p.config.GetMatchLevel())
				}
				if len(filteredSNPs) > 0 { // Only include group if it has SNPs after filtering
				    filteredGroupings[groupName] = filteredSNPs
				}
			}
			topicOutputData.Groupings = filteredGroupings

			if len(topicOutputData.Groupings) == 0 && p.config.GetMatchLevel() != config.MatchLevelNone { // Don't save empty files unless match level is None
			    logger.Info("Skipping empty topic output after filtering", "topicName", topicName, "matchLevel", p.config.GetMatchLevel())
			    continue
			}

			if err := saveTopicOutput(topicOutputData, outPath); err != nil {
				logger.Error(err, "failed to save topic output", "topic", topicName)
				return nil, nil, fmt.Errorf("failed to save topic output %s: %w", topicName, err)
			}
			outputFiles = append(outputFiles, outPath)
			groupFilenames[topicName] = filename // Using topicName as key for consistency in logging
		}

	} else { // Original logic for groupingMode == "group"
		groupings := make(map[string]*models.Grouping)
		for _, record := range records {
			logger.Debug("Processing record for group mode", "record", record)
			if len(record) != 7 {
				errorRecords = append(errorRecords, fmt.Sprintf("Record with %d columns: %v", len(record), record))
				logger.Info("Invalid record format", "record", record)
				continue
			}

			actualTopic := record[0]
			groupKey := record[1] // Group column is the key
			genotype := record[5]

			if _, exists := groupings[groupKey]; !exists {
				newGroup := &models.Grouping{
					Topic: actualTopic,
					Name:  groupKey,
				}
				groupings[groupKey] = newGroup
				logger.Debug("New group created", "groupKey", groupKey)
			}

			snp, err := models.NewSNP(record[2], record[3], record[4], record[6], genotype)
			if err != nil {
				errorRecords = append(errorRecords, fmt.Sprintf("Record validation failed: %v", record))
				logger.Info("Skipping SNP due to validation error", "error", err)
				continue
			}
			groupings[groupKey].SNP = append(groupings[groupKey].SNP, *snp)
			logger.Debug("SNP added to group", "groupKey", groupKey, "snpRSID", snp.RSID)
		}

		for groupName, groupingData := range groupings {
			filename := fmt.Sprintf("%s.json", strings.ReplaceAll(groupName, "/", "-"))
			outPath := filepath.Join(p.outputDir, filename)

			var filteredSNPs []models.SNP
			for _, snp := range groupingData.SNP {
				filteredSNPs = models.AddIfMatch(filteredSNPs, snp, p.config.GetMatchLevel())
			}
			
			if len(filteredSNPs) == 0 && p.config.GetMatchLevel() != config.MatchLevelNone { // Don't save empty files unless match level is None
			    logger.Info("Skipping empty group output after filtering", "groupName", groupName, "matchLevel", p.config.GetMatchLevel())
			    continue
			}

			if err := SaveResult(&models.ConversionResult{Grouping: models.Grouping{
				Topic: groupingData.Topic,
				Name:  groupName,
				SNP:   filteredSNPs,
			}}, outPath); err != nil {
				logger.Error(err, "failed to save grouping", "group", groupName)
				return nil, nil, fmt.Errorf("failed to save grouping %s: %w", groupName, err)
			}
			outputFiles = append(outputFiles, outPath)
			groupFilenames[groupName] = filename
		}
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
