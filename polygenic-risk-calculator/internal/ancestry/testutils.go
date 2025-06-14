package ancestry

import "testing"

// CreateTestAncestry creates an ancestry object for testing with error handling
func CreateTestAncestry(t *testing.T, population, gender string) *Ancestry {
	t.Helper()
	ancestry, err := New(population, gender)
	if err != nil {
		t.Fatalf("Failed to create test ancestry %s_%s: %v", population, gender, err)
	}
	return ancestry
}

// MockRowData creates mock BigQuery row data for testing frequency selection
func MockRowData(data map[string]float64) map[string]interface{} {
	result := make(map[string]interface{})
	for col, freq := range data {
		result[col] = freq
	}
	return result
}

// MockRowDataWithColumns creates mock row data with specified columns and values
func MockRowDataWithColumns(columns []string, values []float64) map[string]interface{} {
	if len(columns) != len(values) {
		panic("columns and values must have same length")
	}

	result := make(map[string]interface{})
	for i, col := range columns {
		result[col] = values[i]
	}
	return result
}

// AllTestCombinations returns all valid ancestry/gender combinations for comprehensive testing
func AllTestCombinations() []struct {
	Population string
	Gender     string
	Code       string
} {
	var combinations []struct {
		Population string
		Gender     string
		Code       string
	}

	populations := GetSupportedPopulations()
	genders := GetSupportedGenders()

	// Add all population/gender combinations
	for _, pop := range populations {
		for _, gender := range genders {
			code := buildInternalCode(pop, gender)
			combinations = append(combinations, struct {
				Population string
				Gender     string
				Code       string
			}{
				Population: pop,
				Gender:     gender,
				Code:       code,
			})
		}
	}

	return combinations
}

// TestAllMappingsExist validates that all ancestry combinations have proper mappings
func TestAllMappingsExist(t *testing.T) {
	t.Helper()
	mappings := getBuiltinMappings()

	for _, combo := range AllTestCombinations() {
		if _, exists := mappings[combo.Code]; !exists {
			t.Errorf("Missing mapping for code: %s (population: %s, gender: %s)",
				combo.Code, combo.Population, combo.Gender)
		}
	}
}
