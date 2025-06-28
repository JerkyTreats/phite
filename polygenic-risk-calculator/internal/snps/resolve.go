package snps

// ResolveSNPs decides which SNP input to use (direct or file), parses if needed, and always cleans/validates.
// Returns the canonical SNP list or error. Used by CLI options system.
import (
	"errors"
	"fmt"
	"strings"
)

var ErrNoSNPsProvided = errors.New("no SNPs provided")

func ResolveSNPs(direct []string, file string) ([]string, error) {
	if len(direct) > 0 {
		return CleanAndValidateSNPs(direct)
	}
	if file != "" {
		rsids, err := ParseSNPsFromFile(file)
		if err != nil {
			return nil, err
		}
		return CleanAndValidateSNPs(rsids)
	}
	return nil, ErrNoSNPsProvided
}

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
