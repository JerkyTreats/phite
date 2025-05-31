package snps

// ResolveSNPs decides which SNP input to use (direct or file), parses if needed, and always cleans/validates.
// Returns the canonical SNP list or error. Used by CLI options system.
import "errors"

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
