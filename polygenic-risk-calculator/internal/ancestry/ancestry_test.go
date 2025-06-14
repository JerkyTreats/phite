package ancestry

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		population string
		gender     string
		wantCode   string
		wantErr    bool
	}{
		{
			name:       "Valid European ancestry",
			population: "EUR",
			gender:     "",
			wantCode:   "EUR",
			wantErr:    false,
		},
		{
			name:       "Valid European male",
			population: "EUR",
			gender:     "MALE",
			wantCode:   "EUR_MALE",
			wantErr:    false,
		},
		{
			name:       "Valid African female",
			population: "AFR",
			gender:     "FEMALE",
			wantCode:   "AFR_FEMALE",
			wantErr:    false,
		},
		{
			name:       "Invalid population",
			population: "INVALID",
			gender:     "",
			wantCode:   "",
			wantErr:    true,
		},
		{
			name:       "Invalid gender",
			population: "EUR",
			gender:     "INVALID",
			wantCode:   "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.population, tt.gender)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Code() != tt.wantCode {
					t.Errorf("New() code = %v, want %v", got.Code(), tt.wantCode)
				}
				if got.Population() != tt.population {
					t.Errorf("New() population = %v, want %v", got.Population(), tt.population)
				}
				if got.Gender() != tt.gender {
					t.Errorf("New() gender = %v, want %v", got.Gender(), tt.gender)
				}
			}
		})
	}
}

func TestAncestry_ColumnPrecedence(t *testing.T) {
	tests := []struct {
		name           string
		population     string
		gender         string
		wantPrecedence []string
	}{
		{
			name:           "European ancestry",
			population:     "EUR",
			gender:         "",
			wantPrecedence: []string{"AF_nfe"},
		},
		{
			name:           "European male",
			population:     "EUR",
			gender:         "MALE",
			wantPrecedence: []string{"AF_nfe_male", "AF_nfe", "AF_male"},
		},
		{
			name:           "African female",
			population:     "AFR",
			gender:         "FEMALE",
			wantPrecedence: []string{"AF_afr_female", "AF_afr", "AF_female"},
		},
		{
			name:           "South Asian male",
			population:     "SAS",
			gender:         "MALE",
			wantPrecedence: []string{"AF_sas_male", "AF_sas", "AF_male"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := New(tt.population, tt.gender)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			got := a.ColumnPrecedence()
			if !reflect.DeepEqual(got, tt.wantPrecedence) {
				t.Errorf("ColumnPrecedence() = %v, want %v", got, tt.wantPrecedence)
			}
		})
	}
}

func TestAncestry_SelectFrequency(t *testing.T) {
	// Create test ancestry
	ancestry, err := New("EUR", "MALE")
	if err != nil {
		t.Fatalf("Failed to create ancestry: %v", err)
	}

	tests := []struct {
		name       string
		rowData    map[string]interface{}
		wantFreq   float64
		wantColumn string
		wantErr    bool
	}{
		{
			name: "First column available",
			rowData: map[string]interface{}{
				"AF_nfe_male": 0.25,
				"AF_nfe":      0.30,
				"AF_male":     0.28,
			},
			wantFreq:   0.25,
			wantColumn: "AF_nfe_male",
			wantErr:    false,
		},
		{
			name: "First column zero, use second",
			rowData: map[string]interface{}{
				"AF_nfe_male": 0.0,
				"AF_nfe":      0.30,
				"AF_male":     0.28,
			},
			wantFreq:   0.30,
			wantColumn: "AF_nfe",
			wantErr:    false,
		},
		{
			name: "Only third column available",
			rowData: map[string]interface{}{
				"AF_nfe_male": 0.0,
				"AF_nfe":      0.0,
				"AF_male":     0.28,
			},
			wantFreq:   0.28,
			wantColumn: "AF_male",
			wantErr:    false,
		},
		{
			name: "No columns available",
			rowData: map[string]interface{}{
				"AF_nfe_male": 0.0,
				"AF_nfe":      0.0,
				"AF_male":     0.0,
			},
			wantFreq:   0.0,
			wantColumn: "",
			wantErr:    true,
		},
		{
			name: "Missing columns",
			rowData: map[string]interface{}{
				"other_column": 0.25,
			},
			wantFreq:   0.0,
			wantColumn: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFreq, gotColumn, err := ancestry.SelectFrequency(tt.rowData)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectFrequency() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFreq != tt.wantFreq {
				t.Errorf("SelectFrequency() freq = %v, want %v", gotFreq, tt.wantFreq)
			}
			if gotColumn != tt.wantColumn {
				t.Errorf("SelectFrequency() column = %v, want %v", gotColumn, tt.wantColumn)
			}
		})
	}
}

func TestIsSupported(t *testing.T) {
	tests := []struct {
		name       string
		population string
		gender     string
		want       bool
	}{
		{"Valid EUR", "EUR", "", true},
		{"Valid EUR_MALE", "EUR", "MALE", true},
		{"Valid AFR_FEMALE", "AFR", "FEMALE", true},
		{"Invalid population", "INVALID", "", false},
		{"Invalid gender", "EUR", "INVALID", false},
		{"All supported populations", "AMI", "", true},
		{"All supported populations", "SAS", "FEMALE", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupported(tt.population, tt.gender); got != tt.want {
				t.Errorf("IsSupported() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSupportedPopulations(t *testing.T) {
	got := GetSupportedPopulations()
	expected := []string{"AFR", "AMR", "ASJ", "EAS", "EUR", "FIN", "SAS", "OTH", "AMI"}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("GetSupportedPopulations() = %v, want %v", got, expected)
	}

	// Verify we have exactly 9 populations
	if len(got) != 9 {
		t.Errorf("GetSupportedPopulations() length = %v, want 9", len(got))
	}
}

func TestGetSupportedGenders(t *testing.T) {
	got := GetSupportedGenders()
	expected := []string{"", "MALE", "FEMALE"}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("GetSupportedGenders() = %v, want %v", got, expected)
	}

	// Verify we have exactly 3 gender options
	if len(got) != 3 {
		t.Errorf("GetSupportedGenders() length = %v, want 3", len(got))
	}
}

func TestBuiltinMappingsComplete(t *testing.T) {
	mappings := getBuiltinMappings()

	// Test that we have all expected combinations
	populations := GetSupportedPopulations()
	genders := GetSupportedGenders()

	expectedCount := len(populations)*len(genders) + 2 // +2 for MALE and FEMALE gender-only
	if len(mappings) != expectedCount {
		t.Errorf("getBuiltinMappings() has %d mappings, expected %d", len(mappings), expectedCount)
	}

	// Test that all combinations exist
	for _, pop := range populations {
		for _, gender := range genders {
			code := buildInternalCode(pop, gender)
			if _, exists := mappings[code]; !exists {
				t.Errorf("Missing mapping for code: %s", code)
			}
		}
	}

	// Test gender-only mappings
	if _, exists := mappings["MALE"]; !exists {
		t.Error("Missing mapping for MALE")
	}
	if _, exists := mappings["FEMALE"]; !exists {
		t.Error("Missing mapping for FEMALE")
	}
}

func TestBuildInternalCode(t *testing.T) {
	tests := []struct {
		name       string
		population string
		gender     string
		want       string
	}{
		{"Combined ancestry", "EUR", "", "EUR"},
		{"Male specific", "EUR", "MALE", "EUR_MALE"},
		{"Female specific", "AFR", "FEMALE", "AFR_FEMALE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildInternalCode(tt.population, tt.gender); got != tt.want {
				t.Errorf("buildInternalCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
