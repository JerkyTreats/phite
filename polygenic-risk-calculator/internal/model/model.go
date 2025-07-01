// Package model defines canonical data structures shared across the polygenic-risk-calculator pipeline.
package model

import (
	"fmt"
)

// GWASSNPRecord represents a single SNP record from GWAS summary statistics.
type GWASSNPRecord struct {
	RSID       string
	RiskAllele string
	Beta       float64
	Trait      string // optional
}

// ValidatedSNP represents a user SNP that has been validated against GWAS data.
type ValidatedSNP struct {
	RSID        string
	Genotype    string
	FoundInGWAS bool
}

// AnnotatedSNP represents a user SNP annotated with GWAS and PRS calculation data.
type AnnotatedSNP struct {
	RSID       string
	Genotype   string
	RiskAllele string
	Beta       float64
	Dosage     int
	Trait      string // optional
}

// ReferenceStats holds population-level statistics for PRS normalization.
type ReferenceStats struct {
	Mean     float64
	Std      float64
	Min      float64
	Max      float64
	Ancestry string
	Trait    string
	Model    string
}

// UserGenotype represents a single SNP in the user's genotype file.
type UserGenotype struct {
	RSID     string
	Genotype string
}

// Variant represents a single variant in a PRS model.
type Variant struct {
	ID           string   // Variant identifier (e.g., rsID or chr:pos:ref:alt)
	Chromosome   string   // Chromosome
	Position     int64    // Position
	Ref          string   // Reference allele
	Alt          string   // Alternative allele
	EffectAllele string   // The allele associated with the effect weight
	OtherAllele  string   // The non-effect allele
	EffectWeight float64  // The weight or beta score of the effect allele
	EffectFreq   *float64 // Optional: Frequency of the effect allele
	BetaValue    *float64 // Optional: Beta value
	BetaCILower  *float64 // Optional: Lower bound of Beta's confidence interval
	BetaCIUpper  *float64 // Optional: Upper bound of Beta's confidence interval
	OddsRatio    *float64 // Optional: Odds Ratio
	ORCILower    *float64 // Optional: Lower bound of OR's confidence interval
	ORCIUpper    *float64 // Optional: Upper bound of OR's confidence interval
	VariantID    *string  // Optional: Variant ID (e.g., chr:pos:ref:alt)
	RSID         *string  // Optional: rsID
}

// PRSModel represents a complete PRS model with its variants.
type PRSModel struct {
	ID       string    // Model identifier
	Trait    string    // Associated trait
	Variants []Variant // List of variants in the model
}

// Validate checks if the PRS model is valid.
func (m *PRSModel) Validate() error {
	if m.ID == "" {
		return fmt.Errorf("model ID cannot be empty")
	}
	if len(m.Variants) == 0 {
		return fmt.Errorf("model must contain at least one variant")
	}

	// Check for duplicate variants
	seen := make(map[string]bool)
	for _, v := range m.Variants {
		if v.ID == "" {
			return fmt.Errorf("variant ID cannot be empty")
		}
		if seen[v.ID] {
			return fmt.Errorf("duplicate variant ID: %s", v.ID)
		}
		seen[v.ID] = true
	}

	return nil
}

// GetEffectSizes returns a map of variant IDs to their effect sizes.
func (m *PRSModel) GetEffectSizes() map[string]float64 {
	effects := make(map[string]float64, len(m.Variants))
	for _, v := range m.Variants {
		effects[v.ID] = v.EffectWeight
	}
	return effects
}
