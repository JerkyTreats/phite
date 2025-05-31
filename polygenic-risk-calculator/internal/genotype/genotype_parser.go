package genotype

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

// UserGenotype represents a user's SNP genotype.
type UserGenotype struct {
	RSID     string // SNP identifier
	Genotype string // e.g., "AA", "CT"
}

// ValidatedSNP represents an SNP that has been checked against user data and GWAS data.
type ValidatedSNP struct {
	RSID        string
	Genotype    string
	FoundInGWAS bool
}

// GWASSNPRecord is a simplified representation of a GWAS SNP record for testing purposes.
type GWASSNPRecord struct {
	RSID       string
	RiskAllele string
}

// ParseGenotypeDataInput holds all the necessary inputs for ParseGenotypeData.
type ParseGenotypeDataInput struct {
	GenotypeFilePath string
	RequestedRSIDs   []string
	GWASData         map[string]GWASSNPRecord // rsid -> GWASSNPRecord
}

// ParseGenotypeDataOutput holds the results of parsing and validation.
type ParseGenotypeDataOutput struct {
	UserGenotypes []UserGenotype
	ValidatedSNPs []ValidatedSNP
	SNPsMissing   []string // rsids not found in user data or GWAS, or non-GACT
}

// ParseGenotypeData parses a user genotype file, validates SNPs, and reports missing ones.
// It autodetects the file format (AncestryDNA or 23andMe).
//
// Inputs:
//   - input: ParseGenotypeDataInput struct containing file path, requested SNPs, and GWAS data.
//
// Outputs:
//   - ParseGenotypeDataOutput struct containing user genotypes, validated SNPs, and missing SNPs.
//   - error: if there's an issue with file reading or critical parsing errors.

func ParseGenotypeData(input ParseGenotypeDataInput) (ParseGenotypeDataOutput, error) {
	requested := make(map[string]struct{})
	for _, rsid := range input.RequestedRSIDs {
		requested[rsid] = struct{}{}
	}

	output := ParseGenotypeDataOutput{}
	userGenos := make(map[string]string)

	f, err := os.Open(input.GenotypeFilePath)
	if err != nil {
		return ParseGenotypeDataOutput{}, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	format := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		cols := strings.Split(line, "\t")
		if format == "" {
			if len(cols) >= 5 && cols[0] == "rsid" && cols[3] == "allele1" {
				format = "ancestry"
				continue
			} else if len(cols) >= 4 && cols[0] == "rsid" && cols[3] == "genotype" {
				format = "23andme"
				continue
			} else {
				return ParseGenotypeDataOutput{}, errors.New("unknown file format")
			}
		}
		if format == "ancestry" && len(cols) >= 5 {
			// rsid, chrom, pos, allele1, allele2
			rsid := cols[0]
			if _, ok := requested[rsid]; ok {
				geno := cols[3] + cols[4]
				userGenos[rsid] = geno
			}
		} else if format == "23andme" && len(cols) >= 4 {
			// rsid, chrom, pos, genotype
			rsid := cols[0]
			if _, ok := requested[rsid]; ok {
				geno := cols[3]
				userGenos[rsid] = geno
			}
		} // else: skip malformed lines
	}

	for rsid := range requested {
		geno, found := userGenos[rsid]
		if found && isValidGenotype(geno) {
			output.UserGenotypes = append(output.UserGenotypes, UserGenotype{RSID: rsid, Genotype: geno})
			foundInGWAS := false
			if _, ok := input.GWASData[rsid]; ok {
				foundInGWAS = true
			}
			output.ValidatedSNPs = append(output.ValidatedSNPs, ValidatedSNP{RSID: rsid, Genotype: geno, FoundInGWAS: foundInGWAS})
		} else {
			output.SNPsMissing = append(output.SNPsMissing, rsid)
		}
	}

	return output, nil
}

func isValidGenotype(geno string) bool {
	if len(geno) != 2 {
		return false
	}
	for _, c := range geno {
		switch c {
		case 'A', 'C', 'G', 'T':
			continue
		default:
			return false
		}
	}
	return true
}
