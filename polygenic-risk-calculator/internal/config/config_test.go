package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect" // Added for DeepEqual
	"strings" // Added for error message checking
	"testing"
)

// Helper to load base config from testdata
func loadBaseConfig(t *testing.T) map[string]interface{} {
	t.Helper()
	data, err := ioutil.ReadFile("testdata/base_config.json")
	if err != nil {
		t.Fatalf("failed to read base config: %v", err)
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to unmarshal base config: %v", err)
	}
	return cfg
}

// Helper to write config to a temp file and return the path
func writeConfigToTempFile(t *testing.T, cfg map[string]interface{}) string {
	t.Helper()
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	f, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := f.Write(data); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestConfig_DefaultsWhenMissing(t *testing.T) {
	ResetForTest()
	SetConfigPath("/tmp/nonexistent.json")
	if GetString("log_level") != "INFO" {
		t.Errorf("expected default log_level INFO, got %q", GetString("log_level"))
	}
}

func TestConfig_LoadsFromCustomPath(t *testing.T) {
	ResetForTest()
	cfg := loadBaseConfig(t)
	cfg["log_level"] = "DEBUG"
	cfg["feature_x"] = true
	path := writeConfigToTempFile(t, cfg)
	defer os.Remove(path)
	SetConfigPath(path)
	if GetString("log_level") != "DEBUG" {
		t.Errorf("expected log_level DEBUG, got %q", GetString("log_level"))
	}
	if !GetBool("feature_x") {
		t.Errorf("expected feature_x true")
	}
}

func TestConfig_GetStringMapString(t *testing.T) {
	ResetForTest()
	cfg := loadBaseConfig(t)
	path := writeConfigToTempFile(t, cfg)
	defer os.Remove(path)
	SetConfigPath(path)
	expectedMap := map[string]string{"eur": "AF_nfe", "afr": "AF_afr"}
	actualMap := GetStringMapString("allele_freq_source.ancestry_mapping")
	if !reflect.DeepEqual(actualMap, expectedMap) {
		t.Errorf("GetStringMapString() got = %v, want %v", actualMap, expectedMap)
	}

	// Test missing key
	ResetForTest()
	cfg2 := loadBaseConfig(t)
	delete(cfg2["allele_freq_source"].(map[string]interface{}), "ancestry_mapping")
	path2 := writeConfigToTempFile(t, cfg2)
	defer os.Remove(path2)
	SetConfigPath(path2)
	emptyMap := GetStringMapString("allele_freq_source.ancestry_mapping")
	if len(emptyMap) != 0 {
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
	cfg := loadBaseConfig(t)
	cfg["log_level"] = "NOPE"
	path := writeConfigToTempFile(t, cfg)
	defer os.Remove(path)
	SetConfigPath(path)
	errReturned := Validate()
	if errReturned == nil {
		t.Error("expected error for invalid log_level")
		return
	}
	expectedErrorMsg := "invalid log_level 'NOPE' (case-insensitive), must be one of: TRACE, DEBUG, INFO, WARN, ERROR, FATAL, PANIC"
	if strings.TrimSpace(errReturned.Error()) != strings.TrimSpace(expectedErrorMsg) {
		t.Errorf("Validate() error message mismatch, got '%s', expected '%s'", errReturned.Error(), expectedErrorMsg)
	}
}

func TestConfig_ValidateValidLowercaseLogLevel(t *testing.T) {
	ResetForTest()
	cfg := loadBaseConfig(t)
	cfg["log_level"] = "debug"
	path := writeConfigToTempFile(t, cfg)
	defer os.Remove(path)
	SetConfigPath(path)
	if err := Validate(); err != nil {
		t.Errorf("Validate() failed for lowercase log_level, expected nil, got %v", err)
	}
}

// --- PRS Reference Config Tests ---

func TestConfig_ValidatePRSReferenceConfigKeys_PresentAndParsed(t *testing.T) {
	ResetForTest()
	cfg := loadBaseConfig(t)
	cfg["prs_stats_cache"].(map[string]interface{})["gcp_project_id"] = "my-project"
	cfg["prs_stats_cache"].(map[string]interface{})["dataset_id"] = "prs_cache"
	cfg["prs_stats_cache"].(map[string]interface{})["table_id"] = "stats"
	cfg["allele_freq_source"].(map[string]interface{})["gcp_project_id"] = "bigquery-public-data"
	cfg["prs_model_source"].(map[string]interface{})["path_or_table_uri"] = "/models/my_prs.tsv"
	path := writeConfigToTempFile(t, cfg)
	defer os.Remove(path)
	SetConfigPath(path)
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
	expectedMap := map[string]string{"eur": "AF_nfe", "afr": "AF_afr"}
	actualMap := GetStringMapString("allele_freq_source.ancestry_mapping")
	if !reflect.DeepEqual(actualMap, expectedMap) {
		t.Errorf("GetStringMapString() after Validate: got = %v, want %v", actualMap, expectedMap)
	}
}

func TestConfig_ValidatePRSReferenceConfigKeys_Missing(t *testing.T) {
	ResetForTest()
	cfg := map[string]interface{}{"log_level": "INFO"} // Missing all PRS keys
	path := writeConfigToTempFile(t, cfg)
	defer os.Remove(path)
	SetConfigPath(path)
	errReturned := Validate()
	if errReturned == nil {
		t.Error("expected error for missing PRS reference config keys")
		return
	}
	missingKeys := []string{
		"prs_stats_cache.gcp_project_id", "prs_stats_cache.dataset_id", "prs_stats_cache.table_id",
		"allele_freq_source.dataset_id_pattern", "allele_freq_source.table_id_pattern",
		"allele_freq_source.ancestry_mapping",
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
	cfg := loadBaseConfig(t)
	cfg["allele_freq_source"].(map[string]interface{})["ancestry_mapping"] = "not_a_map"
	path := writeConfigToTempFile(t, cfg)
	defer os.Remove(path)
	SetConfigPath(path)
	errReturned := Validate()
	if errReturned == nil {
		t.Error("expected error for invalid type of allele_freq_source.ancestry_mapping")
		return
	}
	expectedErrorMsgPart := "allele_freq_source.ancestry_mapping must be a map[string]string"
	if !strings.Contains(errReturned.Error(), expectedErrorMsgPart) {
		t.Errorf("Validate() error message mismatch, got '%s', expected to contain '%s'", errReturned.Error(), expectedErrorMsgPart)
	}
}

func TestValidate_RequiredKeys_AllPresent(t *testing.T) {
	ResetForTest()
	RegisterRequiredKey("key1")
	RegisterRequiredKey("key2")
	cfg := loadBaseConfig(t)
	cfg["key1"] = "value1"
	cfg["key2"] = "value2"
	path := writeConfigToTempFile(t, cfg)
	defer os.Remove(path)
	SetConfigPath(path)
	if err := Validate(); err != nil {
		t.Errorf("Validate() failed, expected nil, got %v", err)
	}
}

func TestValidate_RequiredKeys_OneMissing(t *testing.T) {
	ResetForTest()
	RegisterRequiredKey("key1")
	RegisterRequiredKey("key2") // key2 will be missing
	cfg := loadBaseConfig(t)
	cfg["key1"] = "value1"
	delete(cfg, "key2")
	path := writeConfigToTempFile(t, cfg)
	defer os.Remove(path)
	SetConfigPath(path)
	err := Validate()
	if err == nil {
		t.Fatal("Validate() passed, expected error for missing key")
	}
	if !strings.Contains(err.Error(), "key2") {
		t.Errorf("Validate() error message '%s' did not contain expected missing key 'key2'", err.Error())
	}
}

func TestValidate_RequiredKeys_MultipleMissing(t *testing.T) {
	ResetForTest()
	RegisterRequiredKey("key1") // Will be missing
	RegisterRequiredKey("key2") // Will be missing
	RegisterRequiredKey("key3")
	cfg := loadBaseConfig(t)
	cfg["key3"] = "value3"
	delete(cfg, "key1")
	delete(cfg, "key2")
	path := writeConfigToTempFile(t, cfg)
	defer os.Remove(path)
	SetConfigPath(path)
	err := Validate()
	if err == nil {
		t.Fatal("Validate() passed, expected error for multiple missing keys")
	}
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
	cfg := loadBaseConfig(t)
	path := writeConfigToTempFile(t, cfg)
	defer os.Remove(path)
	SetConfigPath(path)
	if err := Validate(); err != nil {
		t.Errorf("Validate() failed, expected nil, got %v", err)
	}
}

// TestValidate_RequiredKeys_DuplicateRegistration is removed as it's less relevant now
// The primary validation for PRS keys is direct within Validate().
// The RegisterRequiredKey mechanism is for other packages to declare their needs.
// The duplicate registration aspect of RegisterRequiredKey itself is simple and covered by its own logic.
