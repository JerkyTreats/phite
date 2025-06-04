package config

import (
	"os"
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
	f.WriteString(`{"log_level": "DEBUG", "feature_x": true}`)
	f.Close()
	SetConfigPath(f.Name())
	if GetString("log_level") != "DEBUG" {
		t.Errorf("expected log_level DEBUG, got %q", GetString("log_level"))
	}
	if !GetBool("feature_x") {
		t.Errorf("expected feature_x true")
	}
}

func TestConfig_ValidateInvalidLogLevel(t *testing.T) {
	ResetForTest()
	f, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(`{"log_level": "NOPE"}`)
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
	f.WriteString(`{"log_level": "debug"}`)
	f.Close()

	SetConfigPath(f.Name())
	if err := Validate(); err != nil {
		t.Errorf("Validate() failed for lowercase log_level, expected nil, got %v", err)
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
	f.WriteString(`{"log_level": "INFO", "key1": "value1", "key2": "value2"}`)
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
	f.WriteString(`{"log_level": "INFO", "key1": "value1"}`)
	f.Close()

	SetConfigPath(f.Name())
	err = Validate()
	if err == nil {
		t.Fatal("Validate() passed, expected error for missing key")
	}
	expectedErrorMsg := "missing required config keys: [key2]"
	if strings.TrimSpace(err.Error()) != strings.TrimSpace(expectedErrorMsg) {
		t.Errorf("Validate() error message mismatch, got '%s', expected '%s'", err.Error(), expectedErrorMsg)
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
	f.WriteString(`{"log_level": "INFO", "key3": "value3"}`)
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
	// No keys registered
	f, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(`{"log_level": "INFO"}`)
	f.Close()

	SetConfigPath(f.Name())
	if err := Validate(); err != nil {
		t.Errorf("Validate() failed, expected nil, got %v", err)
	}
}

func TestValidate_RequiredKeys_DuplicateRegistration(t *testing.T) {
	ResetForTest()
	RegisterRequiredKey("dup_key")
	RegisterRequiredKey("dup_key") // Register same key again

	f, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(`{"log_level": "INFO", "dup_key": "value_dup"}`)
	f.Close()

	SetConfigPath(f.Name())
	if err := Validate(); err != nil {
		t.Errorf("Validate() failed for duplicate registration, expected nil, got %v", err)
	}

	// Test missing with duplicate registration
	ResetForTest()
	RegisterRequiredKey("dup_key_miss")
	RegisterRequiredKey("dup_key_miss")
	f2, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f2.Name())
	f2.WriteString(`{"log_level": "INFO"}`)
	f2.Close()

	SetConfigPath(f2.Name())
	err = Validate()
	if err == nil {
		t.Fatal("Validate() passed, expected error for missing key with duplicate registration")
	}
	expectedErrorMsg := "missing required config keys: [dup_key_miss]"
	if strings.TrimSpace(err.Error()) != strings.TrimSpace(expectedErrorMsg) {
		t.Errorf("Validate() error message mismatch for duplicate registration, got '%s', expected '%s'", err.Error(), expectedErrorMsg)
	}
}
