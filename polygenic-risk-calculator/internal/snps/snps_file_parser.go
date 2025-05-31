package snps

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
		return nil, fmt.Errorf("unsupported file extension: %s (supported: %v)", ext, supported)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return parser(f)
}

func parseJSON(r io.Reader) ([]string, error) {
	var rsids []string
	dec := json.NewDecoder(r)
	if err := dec.Decode(&rsids); err != nil {
		return nil, errors.New("malformed JSON: " + err.Error())
	}
	for i := range rsids {
		rsids[i] = strings.TrimSpace(rsids[i])
		if rsids[i] == "" {
			return nil, errors.New("empty rsid found in input")
		}
	}
	// Deduplicate
	seen := make(map[string]struct{})
	out := make([]string, 0, len(rsids))
	for _, r := range rsids {
		if _, exists := seen[r]; !exists {
			seen[r] = struct{}{}
			out = append(out, r)
		}
	}
	return out, nil
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
		rsids[i] = strings.TrimSpace(rsids[i])
		if strings.ContainsRune(rsids[i], '\x00') {
			return nil, errors.New("malformed input: null byte found")
		}
	}
	seen := make(map[string]struct{})
	out := make([]string, 0, len(rsids))
	for _, r := range rsids {
		if r == "" {
			return nil, errors.New("empty rsid found in input")
		}
		if _, exists := seen[r]; !exists {
			seen[r] = struct{}{}
			out = append(out, r)
		}
	}
	return out, nil
}

func splitDelimited(line string, sep rune) []string {
	return strings.Split(line, string(sep))
}

// splitCSV splits a CSV line on commas, handling simple cases (no quoted fields)
func splitCSV(line string) []string {
	return strings.Split(line, ",")
}
