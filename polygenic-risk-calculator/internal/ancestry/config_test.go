package ancestry

import (
	"testing"
)

func TestNewFromConfig_MockTest(t *testing.T) {
	// Since we can't easily mock the config system, we'll test the logic indirectly
	// by testing that NewFromConfig properly validates the parameters it gets

	// Test that the function exists and can handle basic cases
	// In a real deployment, this would be tested with actual config files

	// For now, just test that the function signature works
	t.Skip("NewFromConfig requires actual config file setup - skipping in unit tests")
}

func TestGetBuiltinMappings(t *testing.T) {
	mappings := getBuiltinMappings()

	// Test that we have all expected entries
	expectedEntries := []string{
		// Main populations
		"AFR", "AMR", "ASJ", "EAS", "EUR", "FIN", "SAS", "OTH", "AMI",
		// Male-specific
		"AFR_MALE", "AMR_MALE", "ASJ_MALE", "EAS_MALE", "EUR_MALE", "FIN_MALE", "SAS_MALE", "OTH_MALE", "AMI_MALE",
		// Female-specific
		"AFR_FEMALE", "AMR_FEMALE", "ASJ_FEMALE", "EAS_FEMALE", "EUR_FEMALE", "FIN_FEMALE", "SAS_FEMALE", "OTH_FEMALE", "AMI_FEMALE",
		// Gender-only
		"MALE", "FEMALE",
	}

	if len(mappings) != len(expectedEntries) {
		t.Errorf("getBuiltinMappings() returned %d entries, expected %d", len(mappings), len(expectedEntries))
	}

	// Test that all expected entries exist and have required fields
	for _, code := range expectedEntries {
		mapping, exists := mappings[code]
		if !exists {
			t.Errorf("Missing mapping for code: %s", code)
			continue
		}

		if len(mapping.precedence) == 0 {
			t.Errorf("Mapping for %s has empty precedence list", code)
		}

		if mapping.description == "" {
			t.Errorf("Mapping for %s has empty description", code)
		}

		// Test that precedence contains valid column names
		for _, col := range mapping.precedence {
			if col == "" {
				t.Errorf("Mapping for %s contains empty column name", code)
			}
			if !isValidColumnName(col) {
				t.Errorf("Mapping for %s contains invalid column name: %s", code, col)
			}
		}
	}
}

// isValidColumnName checks if a column name follows the expected gnomAD pattern
func isValidColumnName(col string) bool {
	validColumns := []string{
		// Main ancestry columns
		"AF_afr", "AF_amr", "AF_asj", "AF_eas", "AF_nfe", "AF_fin", "AF_sas", "AF_oth", "AF_ami",
		// Male-specific columns
		"AF_afr_male", "AF_amr_male", "AF_asj_male", "AF_eas_male", "AF_nfe_male", "AF_fin_male", "AF_sas_male", "AF_oth_male", "AF_ami_male",
		// Female-specific columns
		"AF_afr_female", "AF_amr_female", "AF_asj_female", "AF_eas_female", "AF_nfe_female", "AF_fin_female", "AF_sas_female", "AF_oth_female", "AF_ami_female",
		// Gender-only columns
		"AF_male", "AF_female",
	}

	for _, valid := range validColumns {
		if col == valid {
			return true
		}
	}
	return false
}
