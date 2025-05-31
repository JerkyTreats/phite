package config

import (
	"os"
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
	if err := Validate(); err == nil {
		t.Error("expected error for invalid log_level")
	}
}
