package reference_model

import (
	"fmt"
	"strings"
)

// Variant represents a single variant in a PRS model.
type Variant struct {
	ID           string   // Variant identifier (e.g., rsID or chr:pos:ref:alt)
	Chromosome   string   // Chromosome
	Position     int64    // Position
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

// Validate checks if the model is valid.
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

// FormatVariantID formats a variant ID from its components.
func FormatVariantID(chrom string, pos int64, ref, alt string) string {
	return fmt.Sprintf("%s:%d:%s:%s", chrom, pos, ref, alt)
}

// ParseVariantID parses a variant ID into its components.
func ParseVariantID(id string) (chrom string, pos int64, ref, alt string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 4 {
		return "", 0, "", "", fmt.Errorf("invalid variant ID format: %s", id)
	}

	var position int64
	if _, err := fmt.Sscanf(parts[1], "%d", &position); err != nil {
		return "", 0, "", "", fmt.Errorf("invalid position in variant ID: %s", id)
	}

	return parts[0], position, parts[2], parts[3], nil
}
