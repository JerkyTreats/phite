// Package models defines the data structures for genetic SNP information and
// conversion results. These structures are used throughout the converter to
// represent genetic data in a structured format that can be easily serialized to
// JSON.
package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/JerkyTreats/PHITE/converter/internal/config"
	"github.com/JerkyTreats/PHITE/converter/pkg/logger"
)

// Subject represents the subject's genetic information for a specific SNP.
// It contains the subject's genotype and the match status compared to the reference
// allele.
type Subject struct {
	// Genotype is a 2-character string representing the subject's genotype
	// Possible values: "AA", "AG", "GG", etc.
	// Special cases: "--" (no data), "" (empty), "0" (unknown)
	Genotype string `json:"Genotype"`

	// Match indicates how the subject's genotype matches the reference allele
	// Possible values: "None", "Partial", "Full"
	Match string `json:"Match"`
}

// NewSubject creates a new Subject with the given genotype and allele.
// Returns an error if the genotype is invalid.
func NewSubject(genotype, allele string) (*Subject, error) {
	if genotype == "--" || genotype == "" {
		logger.Debug("Creating special case Subject", "genotype", genotype)
		return &Subject{
			Genotype: genotype,
			Match:    "None",
		}, nil
	}

	if len(genotype) != 2 {
		return nil, fmt.Errorf("invalid genotype format: %s", genotype)
	}

	validNucleotides := "ATCG"
	if !strings.Contains(validNucleotides, string(genotype[0])) ||
		!strings.Contains(validNucleotides, string(genotype[1])) {
		return nil, fmt.Errorf("invalid nucleotides in genotype: %s", genotype)
	}

	subject := &Subject{
		Genotype: genotype,
		Match:    DetermineMatch(genotype, allele),
	}

	if err := subject.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Subject: %w", err)
	}

	logger.Debug("Created new Subject", "genotype", genotype, "match", subject.Match)
	return subject, nil
}

// Validate validates the Subject's genotype format.
// Returns an error if the genotype is invalid.
// Special cases (--, "") are considered valid.
func (s *Subject) Validate() error {
	logger.Debug("Validating Subject", "genotype", s.Genotype, "match", s.Match)

	// Check if genotype is a valid special case
	if s.Genotype == "--" || s.Genotype == "" {
		return nil
	}

	// Check if genotype is valid format (2 characters)
	if len(s.Genotype) != 2 {
		return fmt.Errorf("invalid genotype format: %s", s.Genotype)
	}

	// Check if genotype contains valid nucleotides
	for _, r := range s.Genotype {
		if !strings.ContainsRune("ATCG", r) {
			return fmt.Errorf("invalid nucleotides in genotype: %s", s.Genotype)
		}
	}

	return nil
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

// NewSNP creates a new SNP with the given parameters and validates them.
// Returns an error if any parameter is invalid.
func NewSNP(gene, rsid, allele, notes, genotype string) (*SNP, error) {
	if gene == "" {
		return nil, fmt.Errorf("gene cannot be empty")
	}

	if rsid == "" {
		return nil, fmt.Errorf("RSID cannot be empty")
	}

	if allele == "" {
		return nil, fmt.Errorf("allele cannot be empty")
	}

	if !strings.HasPrefix(rsid, "rs") {
		return nil, fmt.Errorf("invalid RSID format: %s", rsid)
	}

	if _, err := strconv.Atoi(rsid[2:]); err != nil {
		return nil, fmt.Errorf("invalid RSID format: %s", rsid)
	}

	if !strings.Contains("ATCG", strings.ToUpper(allele)) {
		return nil, fmt.Errorf("invalid allele: %s", allele)
	}

	subject, err := NewSubject(genotype, allele)
	if err != nil {
		return nil, fmt.Errorf("invalid subject: %w", err)
	}

	snp := &SNP{
		Gene:    gene,
		RSID:    rsid,
		Allele:  allele,
		Notes:   notes,
		Subject: *subject,
	}

	// Validate the entire SNP
	if err := snp.Validate(); err != nil {
		return nil, fmt.Errorf("invalid SNP: %w", err)
	}

	logger.Debug("Created new SNP", "rsid", rsid, "genotype", genotype)
	return snp, nil
}

// Validate validates the SNP's fields.
// Returns an error if any field is invalid.
func (s *SNP) Validate() error {
	if s.Gene == "" {
		return fmt.Errorf("gene cannot be empty")
	}

	if s.RSID == "" {
		return fmt.Errorf("RSID cannot be empty")
	}

	// Validate RSID format
	if !strings.HasPrefix(s.RSID, "rs") {
		return fmt.Errorf("invalid RSID format: %s", s.RSID)
	}

	// Check if the part after 'rs' is numeric
	if _, err := strconv.Atoi(s.RSID[2:]); err != nil {
		return fmt.Errorf("invalid RSID format: %s", s.RSID)
	}

	if s.Allele == "" {
		return fmt.Errorf("allele cannot be empty")
	}

	// Validate allele
	validAlleles := "ATCG"
	if !strings.Contains(validAlleles, s.Allele) {
		return fmt.Errorf("invalid allele: %s", s.Allele)
	}

	// Validate subject
	if err := s.Subject.Validate(); err != nil {
		return fmt.Errorf("invalid subject: %w", err)
	}

	// Add more SNP-specific validations if needed
	return nil
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

// ToString returns the Grouping as a pretty-printed JSON string
func (g *Grouping) ToString() string {
	b, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}

// ConversionResult represents the final output of the SNP conversion process.
// It contains a single Grouping of SNPs organized by their biological relationships.
type ConversionResult struct {
	Grouping Grouping `json:"Grouping"`
}

// AddIfMatch appends snp to the slice if it matches the config match level.
func AddIfMatch(snps []SNP, snp SNP, matchLevel config.MatchLevel) []SNP {
	match := DetermineMatch(snp.Subject.Genotype, snp.Allele)
	switch matchLevel {
	case config.MatchLevelNone:
		return append(snps, snp)
	case config.MatchLevelPartial:
		if match == "Partial" || match == "Full" {
			return append(snps, snp)
		}
	case config.MatchLevelFull:
		if match == "Full" {
			return append(snps, snp)
		}
	}
	return snps
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
