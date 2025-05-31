// Package model defines canonical data structures shared across the polygenic-risk-calculator pipeline.
package model

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
