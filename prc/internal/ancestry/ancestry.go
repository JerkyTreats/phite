package ancestry

import (
	"fmt"
	"slices"

	"phite.io/polygenic-risk-calculator/internal/utils"
)

// Ancestry represents a population and gender combination for frequency lookups
type Ancestry struct {
	population  string   // Population code (e.g., "EUR", "AFR")
	gender      string   // Gender code (e.g., "MALE", "FEMALE", "" for combined)
	code        string   // Combined code (e.g., "EUR_FEMALE", "AFR")
	description string   // Full description
	precedence  []string // Column precedence order for frequency selection
}

// New creates a new Ancestry object from separate population and gender codes
func New(population, gender string) (*Ancestry, error) {
	if !IsSupported(population, gender) {
		return nil, fmt.Errorf("unsupported ancestry combination: population=%s, gender=%s", population, gender)
	}

	code := buildInternalCode(population, gender)
	mapping, exists := getBuiltinMappings()[code]
	if !exists {
		return nil, fmt.Errorf("no mapping found for ancestry code: %s", code)
	}

	return &Ancestry{
		population:  population,
		gender:      gender,
		code:        code,
		description: mapping.description,
		precedence:  mapping.precedence,
	}, nil
}

// Code returns the combined ancestry code (e.g., "EUR_MALE", "AFR")
func (a *Ancestry) Code() string {
	return a.code
}

// Population returns the population component
func (a *Ancestry) Population() string {
	return a.population
}

// Gender returns the gender component (empty string for combined)
func (a *Ancestry) Gender() string {
	return a.gender
}

// Description returns the full human-readable description
func (a *Ancestry) Description() string {
	return a.description
}

// ColumnPrecedence returns the ordered list of columns to try for frequency selection
func (a *Ancestry) ColumnPrecedence() []string {
	return a.precedence
}

// SelectFrequency selects the best available frequency from row data using precedence order
func (a *Ancestry) SelectFrequency(rowData map[string]interface{}) (float64, string, error) {
	for _, col := range a.precedence {
		if freq := utils.ToFloat64(rowData[col]); freq > 0 {
			return freq, col, nil // Return first non-zero frequency found
		}
	}
	return 0, "", fmt.Errorf("no frequency data available for %s", a.code)
}

// IsSupported validates if the population and gender combination is supported
func IsSupported(population, gender string) bool {
	return slices.Contains(getSupportedPopulations(), population) &&
		slices.Contains(getSupportedGenders(), gender)
}

// GetSupportedPopulations returns all supported population codes
func GetSupportedPopulations() []string {
	return getSupportedPopulations()
}

// GetSupportedGenders returns all supported gender codes
func GetSupportedGenders() []string {
	return getSupportedGenders()
}

// buildInternalCode combines population and gender into internal code
func buildInternalCode(population, gender string) string {
	if gender == "" {
		return population // "EUR"
	}
	return fmt.Sprintf("%s_%s", population, gender) // "EUR_MALE", "AFR_FEMALE"
}

// getSupportedPopulations returns the list of supported population codes
func getSupportedPopulations() []string {
	return []string{
		"AFR", "AMR", "ASJ", "EAS", "EUR", "FIN", "SAS", "OTH", "AMI",
	}
}

// getSupportedGenders returns the list of supported gender codes
func getSupportedGenders() []string {
	return []string{
		"",       // Combined/default (no gender specification)
		"MALE",   // Male-specific frequencies
		"FEMALE", // Female-specific frequencies
	}
}
