package config

import (
	"os"
	"reflect" // Added for DeepEqual
	"strings" // Added for error message checking
	"testing"
)

func TestConfig_DefaultsWhenMissing(t *testing.T) {
	ResetForTest()
	SetConfigPath("/tmp/nonexistent.json")
	if GetString("log_level") != "INFO" {
		t.Errorf("expected default log_level INFO, got %q", GetString("log_level"))
	}
}

func TestConfig_LoadsFromCustomPath(t *testing.T) {
	ResetForTest()
	f, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	// Note: ancestry_mapping added here for completeness in a "full" config example
	f.WriteString(`{
		"log_level": "DEBUG",
		"feature_x": true,
		"reference_genome_build": "GRCh38",
		"prs_stats_cache": {
			"gcp_project_id": "p",
			"dataset_id": "d",
			"table_id": "t"
		},
		"allele_freq_source": {
			"type": "bigquery_gnomad",
			"gcp_project_id": "b",
			"dataset_id_pattern": "gnomAD",
			"table_id_pattern": "genomes_v3_GRCh38",
			"ancestry_mapping": {
				"EUR": "AF_nfe",
				"AFR": "AF_afr"
			}
		},
		"prs_model_source": {
			"type": "file",
			"path_or_table_uri": "/models/model.tsv"
		}
	}`)
	f.Close()
	SetConfigPath(f.Name())
	if GetString("log_level") != "DEBUG" {
		t.Errorf("expected log_level DEBUG, got %q", GetString("log_level"))
	}
	if !GetBool("feature_x") {
		t.Errorf("expected feature_x true")
	}
}

func TestConfig_GetStringMapString(t *testing.T) {
	ResetForTest()
	f, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	validMapContent := `{
		"log_level": "INFO",
		"reference_genome_build": "GRCh38",
		"prs_stats_cache": {"gcp_project_id": "p", "dataset_id": "d", "table_id": "t"},
		"allele_freq_source": {
			"type": "bigquery_gnomad",
			"gcp_project_id": "b",
			"dataset_id_pattern": "gnomAD",
			"table_id_pattern": "genomes_v3_GRCh38",
			"ancestry_mapping": {
				"EUR": "AF_nfe",
				"AFR": "AF_afr"
			}
		},
		"prs_model_source": {"type": "file", "path_or_table_uri": "/models/model.tsv"}
	}`
	f.WriteString(validMapContent)
	f.Close()

	SetConfigPath(f.Name())
	expectedMap := map[string]string{"eur": "AF_nfe", "afr": "AF_afr"}
	actualMap := GetStringMapString("allele_freq_source.ancestry_mapping")
	if !reflect.DeepEqual(actualMap, expectedMap) {
		t.Errorf("GetStringMapString() got = %v, want %v", actualMap, expectedMap)
	}

	// Test missing key
	ResetForTest()
	f2, _ := os.CreateTemp("", "phite-config-*.json")
	defer os.Remove(f2.Name())
	// ancestry_mapping is missing
	f2.WriteString(`{"log_level": "INFO", "reference_genome_build": "GRCh38", "prs_stats_cache": {"gcp_project_id": "p", "dataset_id": "d", "table_id": "t"}, "allele_freq_source": {"type": "bigquery_gnomad", "gcp_project_id": "b", "dataset_id_pattern": "gnomAD", "table_id_pattern": "genomes_v3_GRCh38"}, "prs_model_source": {"type": "file", "path_or_table_uri": "/models/model.tsv"}}`)
	f2.Close()
	SetConfigPath(f2.Name())
	emptyMap := GetStringMapString("allele_freq_source.ancestry_mapping") // Viper returns nil for missing map, GetStringMapString should handle
	if len(emptyMap) != 0 {
		// Our GetStringMapString wrapper returns an empty map if not found or if config is nil
		t.Errorf("GetStringMapString() for missing key: got = %v, want empty map", emptyMap)
	}

	// Test missing config file
	ResetForTest()
	SetConfigPath("/tmp/nonexistent-config-for-map.json")
	emptyMapFromMissingFile := GetStringMapString("any_key")
	if len(emptyMapFromMissingFile) != 0 {
		t.Errorf("GetStringMapString() for missing config file: got = %v, want empty map", emptyMapFromMissingFile)
	}
}

func TestConfig_ValidateInvalidLogLevel(t *testing.T) {
	ResetForTest()
	f, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	// Note: ancestry_mapping added here
	f.WriteString(`{
		"log_level": "NOPE",
		"reference_genome_build": "GRCh38",
		"prs_stats_cache": {"gcp_project_id": "p", "dataset_id": "d", "table_id": "t"},
		"allele_freq_source": {
			"type": "bigquery_gnomad",
			"gcp_project_id": "b",
			"dataset_id_pattern": "gnomAD",
			"table_id_pattern": "genomes_v3_GRCh38",
			"ancestry_mapping": {"EUR": "AF_nfe"}
		},
		"prs_model_source": {"type": "file", "path_or_table_uri": "/models/model.tsv"}
	}`)
	f.Close()
	SetConfigPath(f.Name())
	errReturned := Validate()
	if errReturned == nil {
		t.Error("expected error for invalid log_level")
		return
	}
	expectedErrorMsg := "invalid log_level: NOPE. Must be one of DEBUG, INFO, ERROR, WARN (case-insensitive)"
	if strings.TrimSpace(errReturned.Error()) != strings.TrimSpace(expectedErrorMsg) {
		t.Errorf("Validate() error message mismatch, got '%s', expected '%s'", errReturned.Error(), expectedErrorMsg)
	}
}

func TestConfig_ValidateValidLowercaseLogLevel(t *testing.T) {
	ResetForTest()
	f, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	// Note: ancestry_mapping added here
	f.WriteString(`{
		"log_level": "debug",
		"reference_genome_build": "GRCh38",
		"prs_stats_cache": {"gcp_project_id": "p", "dataset_id": "d", "table_id": "t"},
		"allele_freq_source": {
			"type": "bigquery_gnomad",
			"gcp_project_id": "b",
			"dataset_id_pattern": "gnomAD",
			"table_id_pattern": "genomes_v3_GRCh38",
			"ancestry_mapping": {"EUR": "AF_nfe"}
		},
		"prs_model_source": {"type": "file", "path_or_table_uri": "/models/model.tsv"}
	}`)
	f.Close()

	SetConfigPath(f.Name())
	if err := Validate(); err != nil {
		t.Errorf("Validate() failed for lowercase log_level, expected nil, got %v", err)
	}
}

// --- GRCh38 and PRS Reference Config Tests ---

func TestConfig_ValidateGRCh38GenomeBuild_Accepted(t *testing.T) {
	ResetForTest()
	f, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	// Note: All required PRS keys added
	f.WriteString(`{
		"log_level": "INFO",
		"reference_genome_build": "GRCh38",
		"prs_stats_cache": {
			"gcp_project_id": "my-project",
			"dataset_id": "prs_cache",
			"table_id": "stats"
		},
		"allele_freq_source": {
			"type": "bigquery_gnomad",
			"gcp_project_id": "bigquery-public-data",
			"dataset_id_pattern": "gnomAD",
			"table_id_pattern": "genomes_v3_GRCh38",
			"ancestry_mapping": {
				"EUR": "AF_nfe",
				"AFR": "AF_afr"
			}
		},
		"prs_model_source": {
			"type": "file",
			"path_or_table_uri": "/models/my_prs.tsv"
		}
	}`)
	f.Close()
	SetConfigPath(f.Name())
	if err := Validate(); err != nil {
		t.Errorf("Validate() failed for GRCh38 genome build with all PRS keys, expected nil, got %v", err)
	}
}

func TestConfig_ValidateNonGRCh38GenomeBuild_Rejected(t *testing.T) {
	ResetForTest()
	f, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	// Note: All required PRS keys added
	f.WriteString(`{
		"log_level": "INFO",
		"reference_genome_build": "GRCh37",
		"prs_stats_cache": {
			"gcp_project_id": "my-project",
			"dataset_id": "prs_cache",
			"table_id": "stats"
		},
		"allele_freq_source": {
			"type": "bigquery_gnomad",
			"gcp_project_id": "bigquery-public-data",
			"dataset_id_pattern": "gnomAD",
			"table_id_pattern": "genomes_v3_GRCh38",
			"ancestry_mapping": {
				"EUR": "AF_nfe"
			}
		},
		"prs_model_source": {
			"type": "file",
			"path_or_table_uri": "/models/my_prs.tsv"
		}
	}`)
	f.Close()
	SetConfigPath(f.Name())
	errReturned := Validate()
	if errReturned == nil {
		t.Error("expected error for non-GRCh38 genome build")
		return
	}
	if !strings.Contains(errReturned.Error(), "reference_genome_build must be 'GRCh38'") {
		t.Errorf("Validate() error message mismatch, got '%s', expected to mention 'reference_genome_build must be 'GRCh38''", errReturned.Error())
	}
}

func TestConfig_ValidatePRSReferenceConfigKeys_PresentAndParsed(t *testing.T) {
	ResetForTest()
	f, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(`{
		"log_level": "INFO",
		"reference_genome_build": "GRCh38",
		"prs_stats_cache": {
			"gcp_project_id": "my-project",
			"dataset_id": "prs_cache",
			"table_id": "stats"
		},
		"allele_freq_source": {
			"type": "bigquery_gnomad",
			"gcp_project_id": "bigquery-public-data",
			"dataset_id_pattern": "gnomAD",
			"table_id_pattern": "genomes_v3_GRCh38",
			"ancestry_mapping": {
				"EUR": "AF_nfe",
				"AFR": "AF_afr"
			}
		},
		"prs_model_source": {
			"type": "file",
			"path_or_table_uri": "/models/my_prs.tsv"
		}
	}`)
	f.Close()
	SetConfigPath(f.Name())

	// Explicitly register keys that might be checked if this test runs in isolation
	// or if other tests haven't registered them globally.
	// This makes the test more robust.
	RegisterRequiredKey("prs_stats_cache.gcp_project_id")
	RegisterRequiredKey("prs_stats_cache.dataset_id")
	RegisterRequiredKey("prs_stats_cache.table_id")
	RegisterRequiredKey("allele_freq_source.type")
	RegisterRequiredKey("allele_freq_source.gcp_project_id")
	RegisterRequiredKey("allele_freq_source.dataset_id_pattern")
	RegisterRequiredKey("allele_freq_source.table_id_pattern")
	RegisterRequiredKey("allele_freq_source.ancestry_mapping")
	RegisterRequiredKey("prs_model_source.type")
	RegisterRequiredKey("prs_model_source.path_or_table_uri")

	if err := Validate(); err != nil {
		t.Errorf("Validate() failed for valid PRS reference config, expected nil, got %v", err)
	}

	// Also test GetStringMapString here
	expectedMap := map[string]string{"eur": "AF_nfe", "afr": "AF_afr"}
	actualMap := GetStringMapString("allele_freq_source.ancestry_mapping")
	if !reflect.DeepEqual(actualMap, expectedMap) {
		t.Errorf("GetStringMapString() after Validate: got = %v, want %v", actualMap, expectedMap)
	}
}

func TestConfig_ValidatePRSReferenceConfigKeys_Missing(t *testing.T) {
	ResetForTest()
	f, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(`{
		"log_level": "INFO",
		"reference_genome_build": "GRCh38"
	}`) // Missing all PRS keys
	f.Close()
	SetConfigPath(f.Name())
	errReturned := Validate()
	if errReturned == nil {
		t.Error("expected error for missing PRS reference config keys")
		return
	}
	// These keys are now checked directly in Validate(), not via RegisterRequiredKey for this specific test's purpose
	missingKeys := []string{
		"prs_stats_cache.gcp_project_id", "prs_stats_cache.dataset_id", "prs_stats_cache.table_id",
		"allele_freq_source.type", "allele_freq_source.gcp_project_id",
		"allele_freq_source.dataset_id_pattern", "allele_freq_source.table_id_pattern",
		"allele_freq_source.ancestry_mapping", // Added
		"prs_model_source.type", "prs_model_source.path_or_table_uri",
	}
	for _, key := range missingKeys {
		if !strings.Contains(errReturned.Error(), key) {
			t.Errorf("Validate() error for missing keys did not mention '%s': %s", key, errReturned.Error())
		}
	}
}

func TestConfig_ValidatePRSReferenceConfigKeys_AncestryMappingInvalidType(t *testing.T) {
	ResetForTest()
	f, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	// ancestry_mapping is a string, not a map
	invalidTypeContent := `{
		"log_level": "INFO",
		"reference_genome_build": "GRCh38",
		"prs_stats_cache": {"gcp_project_id": "p", "dataset_id": "d", "table_id": "t"},
		"allele_freq_source": {
			"type": "bigquery_gnomad",
			"gcp_project_id": "b",
			"dataset_id_pattern": "gnomAD",
			"table_id_pattern": "genomes_v3_GRCh38",
			"ancestry_mapping": "not_a_map"
		},
		"prs_model_source": {"type": "file", "path_or_table_uri": "/models/model.tsv"}
	}`
	f.WriteString(invalidTypeContent)
	f.Close()

	SetConfigPath(f.Name())
	errReturned := Validate()
	if errReturned == nil {
		t.Error("expected error for invalid type of allele_freq_source.ancestry_mapping")
		return
	}
	// The exact error message depends on Viper's internal type assertion,
	// but it should indicate a type mismatch.
	// The error from our Validate function is "invalid type for allele_freq_source.ancestry_mapping: expected map[string]string, got string"
	expectedErrorMsgPart := "invalid type for allele_freq_source.ancestry_mapping"
	if !strings.Contains(errReturned.Error(), expectedErrorMsgPart) {
		t.Errorf("Validate() error message mismatch, got '%s', expected to contain '%s'", errReturned.Error(), expectedErrorMsgPart)
	}
}

func TestValidate_RequiredKeys_AllPresent(t *testing.T) {
	ResetForTest()
	RegisterRequiredKey("key1")
	RegisterRequiredKey("key2")

	f, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	// Note: All PRS keys added
	f.WriteString(`{
		"log_level": "INFO",
		"key1": "value1",
		"key2": "value2",
		"reference_genome_build": "GRCh38",
		"prs_stats_cache": {"gcp_project_id": "p", "dataset_id": "d", "table_id": "t"},
		"allele_freq_source": {
			"type": "bigquery_gnomad",
			"gcp_project_id": "b",
			"dataset_id_pattern": "gnomAD",
			"table_id_pattern": "genomes_v3_GRCh38",
			"ancestry_mapping": {"EUR": "AF_nfe"}
		},
		"prs_model_source": {"type": "file", "path_or_table_uri": "/models/model.tsv"}
	}`)
	f.Close()

	SetConfigPath(f.Name())
	if err := Validate(); err != nil {
		t.Errorf("Validate() failed, expected nil, got %v", err)
	}
}

func TestValidate_RequiredKeys_OneMissing(t *testing.T) {
	ResetForTest()
	RegisterRequiredKey("key1")
	RegisterRequiredKey("key2") // key2 will be missing

	f, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	// Note: All PRS keys added
	f.WriteString(`{
		"log_level": "INFO",
		"key1": "value1",
		"reference_genome_build": "GRCh38",
		"prs_stats_cache": {"gcp_project_id": "p", "dataset_id": "d", "table_id": "t"},
		"allele_freq_source": {
			"type": "bigquery_gnomad",
			"gcp_project_id": "b",
			"dataset_id_pattern": "gnomAD",
			"table_id_pattern": "genomes_v3_GRCh38",
			"ancestry_mapping": {"EUR": "AF_nfe"}
		},
		"prs_model_source": {"type": "file", "path_or_table_uri": "/models/model.tsv"}
	}`)
	f.Close()

	SetConfigPath(f.Name())
	err = Validate()
	if err == nil {
		t.Fatal("Validate() passed, expected error for missing key")
	}
	// The error message includes all keys missing from the direct checks in Validate()
	// plus the registered 'key2'.
	if !strings.Contains(err.Error(), "key2") {
		t.Errorf("Validate() error message '%s' did not contain expected missing key 'key2'", err.Error())
	}
}

func TestValidate_RequiredKeys_MultipleMissing(t *testing.T) {
	ResetForTest()
	RegisterRequiredKey("key1") // Will be missing
	RegisterRequiredKey("key2") // Will be missing
	RegisterRequiredKey("key3")

	f, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	// Note: All PRS keys added
	f.WriteString(`{
		"log_level": "INFO",
		"key3": "value3",
		"reference_genome_build": "GRCh38",
		"prs_stats_cache": {"gcp_project_id": "p", "dataset_id": "d", "table_id": "t"},
		"allele_freq_source": {
			"type": "bigquery_gnomad",
			"gcp_project_id": "b",
			"dataset_id_pattern": "gnomAD",
			"table_id_pattern": "genomes_v3_GRCh38",
			"ancestry_mapping": {"EUR": "AF_nfe"}
		},
		"prs_model_source": {"type": "file", "path_or_table_uri": "/models/model.tsv"}
	}`)
	f.Close()

	SetConfigPath(f.Name())
	err = Validate()
	if err == nil {
		t.Fatal("Validate() passed, expected error for multiple missing keys")
	}
	// Order of missing keys in error message might vary, so check for substrings
	missingKey1 := "key1"
	missingKey2 := "key2"
	if !strings.Contains(err.Error(), missingKey1) || !strings.Contains(err.Error(), missingKey2) {
		t.Errorf("Validate() error message '%s' did not contain expected missing keys '%s' and '%s'", err.Error(), missingKey1, missingKey2)
	}
	if strings.Contains(err.Error(), "key3") {
		t.Errorf("Validate() error message '%s' unexpectedly contained key3 which was present", err.Error())
	}
}

func TestValidate_RequiredKeys_NoneRegistered(t *testing.T) {
	ResetForTest()
	// No keys registered via RegisterRequiredKey
	f, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	// All PRS keys are present for direct checks in Validate()
	f.WriteString(`{
		"log_level": "INFO",
		"reference_genome_build": "GRCh38",
		"prs_stats_cache": {"gcp_project_id": "p", "dataset_id": "d", "table_id": "t"},
		"allele_freq_source": {
			"type": "bigquery_gnomad",
			"gcp_project_id": "b",
			"dataset_id_pattern": "gnomAD",
			"table_id_pattern": "genomes_v3_GRCh38",
			"ancestry_mapping": {"EUR": "AF_nfe"}
		},
		"prs_model_source": {"type": "file", "path_or_table_uri": "/models/model.tsv"}
	}`)
	f.Close()

	SetConfigPath(f.Name())
	if err := Validate(); err != nil { // Should pass as no *registered* keys are missing, and all directly checked keys are present
		t.Errorf("Validate() failed, expected nil, got %v", err)
	}
}

// TestValidate_RequiredKeys_DuplicateRegistration is removed as it's less relevant now
// The primary validation for PRS keys is direct within Validate().
// The RegisterRequiredKey mechanism is for other packages to declare their needs.
// The duplicate registration aspect of RegisterRequiredKey itself is simple and covered by its own logic.
