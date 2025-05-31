package snps

import (
	"fmt"
	"strings"
)

// CleanAndValidateSNPs trims, deduplicates, and validates a list of SNP rsids.
// Returns a cleaned list or an error if any rsid is empty or invalid.
func CleanAndValidateSNPs(rsids []string) ([]string, error) {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(rsids))
	for _, r := range rsids {
		r = strings.TrimSpace(r)
		if r == "" {
			return nil, fmt.Errorf("empty rsid in SNP list")
		}
		if _, exists := seen[r]; !exists {
			seen[r] = struct{}{}
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no SNPs provided")
	}
	return out, nil
}
