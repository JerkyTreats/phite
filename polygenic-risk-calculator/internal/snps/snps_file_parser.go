package snps

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"phite.io/polygenic-risk-calculator/internal/logging"
)

// ParseSNPsFromFile parses a list of rsids from a file (CSV or JSON).
// Returns a deduplicated, trimmed slice of rsids or an error.
func ParseSNPsFromFile(path string) ([]string, error) {
	ext := strings.ToLower(filepath.Ext(path))

	parsers := map[string]func(io.Reader) ([]string, error){
		".json": parseJSON,
		".csv":  parseCSV,
		".tsv":  parseTSV,
	}

	parser, ok := parsers[ext]
	if !ok {
		supported := make([]string, 0, len(parsers))
		for k := range parsers {
			supported = append(supported, k)
		}
		logging.Error("unsupported file extension: %s (supported: %v)", ext, supported)
		return nil, errors.New("unsupported file extension")
	}

	logging.Info("Opening SNP file: %s", path)
	f, err := os.Open(path)
	if err != nil {
		logging.Error("failed to open SNP file: %s, err: %v", path, err)
		return nil, err
	}
	defer f.Close()
	logging.Info("Detected SNP file format: %s", ext)
	out, err := parser(f)
	if err != nil {
		logging.Error("failed to parse SNP file: %s, err: %v", path, err)
		return nil, err
	}
	logging.Info("Parsed %d SNP rsids from file %s", len(out), path)
	return out, nil
}

func parseJSON(r io.Reader) ([]string, error) {
	var rsids []string
	dec := json.NewDecoder(r)
	if err := dec.Decode(&rsids); err != nil {
		logging.Error("malformed JSON: %v", err)
		return nil, errors.New("malformed JSON: " + err.Error())
	}
	rsids, err := CleanAndValidateSNPs(rsids)
	if err != nil {
		logging.Error("invalid rsids in JSON input: %v", err)
		return nil, err
	}
	return rsids, nil
}

func parseCSV(r io.Reader) ([]string, error) {
	return parseDelimited(r, ',')
}

func parseTSV(r io.Reader) ([]string, error) {
	return parseDelimited(r, '\t')
}

func parseDelimited(r io.Reader, sep rune) ([]string, error) {
	rsids := []string{}
	scanner := bufio.NewScanner(r)
	var rsidColIdx int = -1
	headerParsed := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !headerParsed {
			headerParsed = true
			fields := splitDelimited(line, sep)
			if len(fields) == 1 {
				if strings.EqualFold(fields[0], "rsid") {
					continue // skip header
				}
				rsids = append(rsids, fields[0])
				continue
			}
			for i, name := range fields {
				norm := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), " ", ""))
				if norm == "rsid" {
					rsidColIdx = i
					break
				}
			}
			if rsidColIdx == -1 {
				logging.Error("header does not contain 'rsid' column in delimited SNP file")
				return nil, errors.New("header does not contain 'rsid' column")
			}
			continue
		}
		fields := splitDelimited(line, sep)
		if rsidColIdx != -1 {
			if rsidColIdx >= len(fields) {
				// Skip short rows (e.g., due to trailing header columns or malformed data)
				continue
			}
			rsids = append(rsids, fields[rsidColIdx])
		} else {
			rsids = append(rsids, fields[0])
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	for i := range rsids {
		if strings.ContainsRune(rsids[i], '\x00') {
			logging.Error("malformed input: null byte found in SNP file")
			return nil, errors.New("malformed input: null byte found")
		}
	}
	rsids, err := CleanAndValidateSNPs(rsids)
	if err != nil {
		logging.Error("invalid rsids in delimited SNP file input: %v", err)
		return nil, err
	}
	return rsids, nil
}

func splitDelimited(line string, sep rune) []string {
	return strings.Split(line, string(sep))
}
