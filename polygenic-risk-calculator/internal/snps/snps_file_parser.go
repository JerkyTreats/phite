package snps

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ParseSNPsFromFile parses a list of rsids from a file (CSV or JSON).
// Returns a deduplicated, trimmed slice of rsids or an error.
func ParseSNPsFromFile(path string) ([]string, error) {
	ext := filepath.Ext(path)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	switch ext {
	case ".json":
		var rsids []string
		if err := json.NewDecoder(f).Decode(&rsids); err != nil {
			return nil, err
		}
		// SNPS_TRIM_JSON
		for i := range rsids {
			rsids[i] = strings.TrimSpace(rsids[i])
		}
		// Deduplicate and validate
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
	case ".csv":
		var rsids []string
		scanner := bufio.NewScanner(f)
		var rsidColIdx int = -1
		headerParsed := false
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			// Parse header if not done
			if !headerParsed {
				headerParsed = true
				fields := splitCSV(line)
				if len(fields) == 1 {
					// Single column: treat as header if "rsid", else treat as data
					if strings.EqualFold(fields[0], "rsid") {
						continue // skip header
					}
					// No header, treat as data
					rsids = append(rsids, fields[0])
					continue
				}
				// Multi-column: look for rsid column
				for i, name := range fields {
					if strings.EqualFold(strings.TrimSpace(name), "rsid") {
						rsidColIdx = i
						break
					}
				}
				if rsidColIdx == -1 {
					return nil, errors.New("CSV header does not contain 'rsid' column")
				}
				continue // header parsed, next lines are data
			}
			// Data rows
			fields := splitCSV(line)
			if rsidColIdx != -1 {
				if rsidColIdx >= len(fields) {
					return nil, errors.New("row missing rsid column")
				}
				rsids = append(rsids, fields[rsidColIdx])
			} else {
				rsids = append(rsids, fields[0])
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		// SNPS_TRIM_CSV
		for i := range rsids {
			rsids[i] = strings.TrimSpace(rsids[i])
			if strings.ContainsRune(rsids[i], '\x00') {
				return nil, errors.New("malformed CSV: null byte found")
			}
		}
		// Deduplicate and validate
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
	default:
		return nil, errors.New("unsupported file extension")
	}
}

// splitCSV splits a CSV line on commas, handling simple cases (no quoted fields)
func splitCSV(line string) []string {
	return strings.Split(line, ",")
}
