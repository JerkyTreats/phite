// Package models defines the data structures for genetic SNP information and
// conversion results. These structures are used throughout the converter to
// represent genetic data in a structured format that can be easily serialized to
// JSON.
package models

// Subject represents the subject's genetic information for a specific SNP.
// It contains the subject's genotype and the match status compared to the reference
// allele.
type Subject struct {
	// Genotype is the subject's genotype (e.g., "AA", "AG", "GG")
	Genotype string `json:"Genotype"`
	// Match indicates how the subject's genotype matches the reference allele
	// Possible values: "None", "Partial", "Full"
	Match string `json:"Match"`
}

// SNP represents a single Single Nucleotide Polymorphism (SNP) with its
// associated genetic information.
type SNP struct {
	// Gene is the name of the gene containing the SNP
	Gene string `json:"Gene"`
	// RSID is the reference SNP identifier (e.g., "rs1801133")
	RSID string `json:"RSID"`
	// Allele is the reference allele for this SNP (e.g., "A", "G")
	Allele string `json:"Allele"`
	// Notes contains additional information about the SNP's effects or significance
	Notes string `json:"Notes"`
	// Subject contains the subject's specific information for this SNP
	Subject Subject `json:"Subject"`
}

// Grouping represents a collection of related SNPs grouped by a common topic.
// SNPs are grouped based on their biological or functional relationship.
type Grouping struct {
	// Topic is the general category or biological function of the SNPs in this group
	Topic string `json:"Topic"`
	// Name is the specific grouping name (e.g., "MTHFR")
	Name string `json:"Name"`
	// SNP is the list of SNPs belonging to this group
	SNP []SNP `json:"SNP"`
}

// ConversionResult represents the final output of the SNP conversion process.
// It contains multiple Groupings of SNPs organized by their biological relationships.
type ConversionResult struct {
	// Groupings is the list of SNP groupings in the result
	Groupings []Grouping `json:"Groupings"`
}

// DetermineMatch determines the match type between a subject's genotype and
// the reference allele.
//
// Args:
//
//	genotype: The subject's genotype (e.g., "AA", "AG", "GG")
//	allele: The reference allele (e.g., "A", "G")
//
// Returns:
//
//	string: The match type, which can be one of:
//	  - "None": No match or invalid genotype
//	  - "Partial": One allele matches the reference
//	  - "Full": Both alleles match the reference
//
// Examples:
//
//	DetermineMatch("AA", "A")  // Returns "Full"
//	DetermineMatch("AG", "A")  // Returns "Partial"
//	DetermineMatch("GG", "A")  // Returns "None"
//	DetermineMatch("--", "A")  // Returns "None"
func DetermineMatch(genotype, allele string) string {
	if genotype == "--" || genotype == "" {
		return "None"
	}

	// Check for full match (both alleles match reference)
	if genotype == allele+allele {
		return "Full"
	}

	// Check for partial match (one allele matches reference)
	if genotype[0] == allele[0] || genotype[1] == allele[0] {
		return "Partial"
	}

	return "None"
}
