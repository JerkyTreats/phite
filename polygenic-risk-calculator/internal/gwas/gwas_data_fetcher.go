package gwas

// Canonical structs from data_model.md
type ValidatedSNP struct {
	RSID        string
	Genotype    string
	FoundInGWAS bool
}

type GWASSNPRecord struct {
	RSID       string
	RiskAllele string
	Beta       float64
	Trait      string // optional
}

type AnnotatedSNP struct {
	RSID       string
	Genotype   string
	RiskAllele string
	Beta       float64
	Dosage     int
	Trait      string // optional
}

type GWASDataFetcherInput struct {
	ValidatedSNPs     []ValidatedSNP
	AssociationsClean []GWASSNPRecord
}

type GWASDataFetcherOutput struct {
	AnnotatedSNPs []AnnotatedSNP
	GWASRecords   []GWASSNPRecord
}

// FetchAndAnnotateGWAS fetches GWAS associations for validated SNPs and annotates them with risk allele, effect size, and computed dosage.
func FetchAndAnnotateGWAS(input GWASDataFetcherInput) GWASDataFetcherOutput {
	var result GWASDataFetcherOutput
	for _, snp := range input.ValidatedSNPs {
		if !snp.FoundInGWAS {
			continue // skip SNPs not found in GWAS
		}
		found := false
		for _, assoc := range input.AssociationsClean {
			if assoc.RSID == snp.RSID {
				found = true
				dosage := computeDosage(snp.Genotype, assoc.RiskAllele)
				annotated := AnnotatedSNP{
					RSID:       snp.RSID,
					Genotype:   snp.Genotype,
					RiskAllele: assoc.RiskAllele,
					Beta:       assoc.Beta,
					Dosage:     dosage,
					Trait:      assoc.Trait,
				}
				result.AnnotatedSNPs = append(result.AnnotatedSNPs, annotated)
				result.GWASRecords = append(result.GWASRecords, assoc)
			}
		}
		if !found {
			// No GWAS association found, skip or handle as needed (here: skip)
		}
	}
	return result
}

// computeDosage calculates the count of the risk allele in the genotype string. Ambiguous/missing genotypes yield 0.
func computeDosage(genotype string, riskAllele string) int {
	if len(genotype) != 2 {
		return 0
	}
	if riskAllele == "" || genotype == "NN" || genotype == "--" {
		return 0
	}
	count := 0
	for i := 0; i < 2; i++ {
		if string(genotype[i]) == riskAllele {
			count++
		}
	}
	return count
}
