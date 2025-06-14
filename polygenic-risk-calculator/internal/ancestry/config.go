package ancestry

import (
	"fmt"

	"phite.io/polygenic-risk-calculator/internal/config"
)

// ancestryInfo holds the mapping information for each ancestry/gender combination
type ancestryInfo struct {
	precedence  []string // Column precedence order for frequency selection
	description string   // Human-readable description
}

// NewFromConfig creates a new Ancestry object from configuration values
func NewFromConfig() (*Ancestry, error) {
	population := config.GetString("ancestry.population")
	gender := config.GetString("ancestry.gender") // defaults to ""

	if population == "" {
		return nil, fmt.Errorf("ancestry.population is required in configuration")
	}

	if !IsSupported(population, gender) {
		return nil, fmt.Errorf("unsupported ancestry combination: population=%s, gender=%s", population, gender)
	}

	return New(population, gender)
}

// getBuiltinMappings returns the hardcoded ancestry mappings based on gnomAD v3 schema
func getBuiltinMappings() map[string]ancestryInfo {
	return map[string]ancestryInfo{
		// Main ancestry populations (single column, no precedence needed)
		"AFR": {precedence: []string{"AF_afr"}, description: "African-American/African ancestry"},
		"AMR": {precedence: []string{"AF_amr"}, description: "Latino ancestry"},
		"ASJ": {precedence: []string{"AF_asj"}, description: "Ashkenazi Jewish ancestry"},
		"EAS": {precedence: []string{"AF_eas"}, description: "East Asian ancestry"},
		"EUR": {precedence: []string{"AF_nfe"}, description: "European ancestry (Non-Finnish European)"},
		"FIN": {precedence: []string{"AF_fin"}, description: "Finnish ancestry"},
		"SAS": {precedence: []string{"AF_sas"}, description: "South Asian ancestry"},
		"OTH": {precedence: []string{"AF_oth"}, description: "Other ancestry"},
		"AMI": {precedence: []string{"AF_ami"}, description: "Amish ancestry"},

		// Male-specific frequencies with precedence order
		"AFR_MALE": {precedence: []string{"AF_afr_male", "AF_afr", "AF_male"}, description: "African-American/African ancestry (males only)"},
		"AMR_MALE": {precedence: []string{"AF_amr_male", "AF_amr", "AF_male"}, description: "Latino ancestry (males only)"},
		"ASJ_MALE": {precedence: []string{"AF_asj_male", "AF_asj", "AF_male"}, description: "Ashkenazi Jewish ancestry (males only)"},
		"EAS_MALE": {precedence: []string{"AF_eas_male", "AF_eas", "AF_male"}, description: "East Asian ancestry (males only)"},
		"EUR_MALE": {precedence: []string{"AF_nfe_male", "AF_nfe", "AF_male"}, description: "European ancestry (males only)"},
		"FIN_MALE": {precedence: []string{"AF_fin_male", "AF_fin", "AF_male"}, description: "Finnish ancestry (males only)"},
		"SAS_MALE": {precedence: []string{"AF_sas_male", "AF_sas", "AF_male"}, description: "South Asian ancestry (males only)"},
		"OTH_MALE": {precedence: []string{"AF_oth_male", "AF_oth", "AF_male"}, description: "Other ancestry (males only)"},
		"AMI_MALE": {precedence: []string{"AF_ami_male", "AF_ami", "AF_male"}, description: "Amish ancestry (males only)"},

		// Female-specific frequencies with precedence order
		"AFR_FEMALE": {precedence: []string{"AF_afr_female", "AF_afr", "AF_female"}, description: "African-American/African ancestry (females only)"},
		"AMR_FEMALE": {precedence: []string{"AF_amr_female", "AF_amr", "AF_female"}, description: "Latino ancestry (females only)"},
		"ASJ_FEMALE": {precedence: []string{"AF_asj_female", "AF_asj", "AF_female"}, description: "Ashkenazi Jewish ancestry (females only)"},
		"EAS_FEMALE": {precedence: []string{"AF_eas_female", "AF_eas", "AF_female"}, description: "East Asian ancestry (females only)"},
		"EUR_FEMALE": {precedence: []string{"AF_nfe_female", "AF_nfe", "AF_female"}, description: "European ancestry (females only)"},
		"FIN_FEMALE": {precedence: []string{"AF_fin_female", "AF_fin", "AF_female"}, description: "Finnish ancestry (females only)"},
		"SAS_FEMALE": {precedence: []string{"AF_sas_female", "AF_sas", "AF_female"}, description: "South Asian ancestry (females only)"},
		"OTH_FEMALE": {precedence: []string{"AF_oth_female", "AF_oth", "AF_female"}, description: "Other ancestry (females only)"},
		"AMI_FEMALE": {precedence: []string{"AF_ami_female", "AF_ami", "AF_female"}, description: "Amish ancestry (females only)"},

		// Gender-combined (ancestry-agnostic, single column)
		"MALE":   {precedence: []string{"AF_male"}, description: "All males (ancestry-combined)"},
		"FEMALE": {precedence: []string{"AF_female"}, description: "All females (ancestry-combined)"},
	}
}

// init registers the required configuration keys
func init() {
	// Register configuration keys with the config system
	config.RegisterRequiredKey("ancestry.population")
	// Note: ancestry.gender is optional and defaults to empty string if not provided
}
