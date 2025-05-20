package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()
	if config.GetLogLevel() != "info" {
		t.Errorf("Expected log level 'info', got '%s'", config.GetLogLevel())
	}
	if config.GetOutputDir() != "output" {
		t.Errorf("Expected output dir 'output', got '%s'", config.GetOutputDir())
	}

	// Test default match level
	if config.GetMatchLevel() != MatchLevelNone {
		t.Errorf("Expected default match level 'None', got '%s'", config.GetMatchLevel())
	}
}

func TestSetMatchLevel(t *testing.T) {
	cfg := NewConfig().(*DefaultConfig)
	cfg.MatchLevel = MatchLevelPartial
	if cfg.GetMatchLevel() != MatchLevelPartial {
		t.Errorf("Expected match level 'Partial', got '%s'", cfg.GetMatchLevel())
	}
	cfg.MatchLevel = MatchLevelFull
	if cfg.GetMatchLevel() != MatchLevelFull {
		t.Errorf("Expected match level 'Full', got '%s'", cfg.GetMatchLevel())
	}
}

func TestSetOutputDir(t *testing.T) {
	config := NewConfig()
	config.SetOutputDir("new_output")
	if config.GetOutputDir() != "new_output" {
		t.Errorf("Expected output dir 'new_output', got '%s'", config.GetOutputDir())
	}

	// Test tilde expansion
	config.SetOutputDir("~/test")
	expectedPath := filepath.Join(os.Getenv("HOME"), "test")
	if config.GetOutputDir() != expectedPath {
		t.Errorf("Expected expanded path '%s', got '%s'", expectedPath, config.GetOutputDir())
	}
}

func TestSaveLoadConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_config_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override HOME for this test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create and save config
	config := NewConfig().(*DefaultConfig)
	config.SetOutputDir("test_output")
	config.MatchLevel = MatchLevelPartial
	if err := config.Save(); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load the config
	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded values
	if loadedConfig.GetOutputDir() != "test_output" {
		t.Errorf("Expected output dir 'test_output', got '%s'", loadedConfig.GetOutputDir())
	}
	if loadedConfig.GetMatchLevel() != MatchLevelPartial {
		t.Errorf("Expected match level 'Partial', got '%s'", loadedConfig.GetMatchLevel())
	}

	// Verify config file exists
	configDir := filepath.Join(tempDir, ".phite")
	configPath := filepath.Join(configDir, "config.json")
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("Config file not found at %s: %v", configPath, err)
	}
}

func TestLoadConfigNonexistent(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_config_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override HOME for this test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Load config that doesn't exist
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Unexpected error loading non-existent config: %v", err)
	}

	// Verify defaults are used
	if config.GetLogLevel() != "info" {
		t.Errorf("Expected default log level 'info', got '%s'", config.GetLogLevel())
	}
	if config.GetOutputDir() != "output" {
		t.Errorf("Expected default output dir 'output', got '%s'", config.GetOutputDir())
	}
}
