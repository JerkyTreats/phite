package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds application configuration
//go:generate mockery --name Config --output mocks

type Config interface {
	GetLogLevel() string
	GetOutputDir() string
	SetOutputDir(dir string)
	Save() error
}

// DefaultConfig implements the Config interface
//go:generate mockery --name DefaultConfig --output mocks

type DefaultConfig struct {
	LogLevel  string `json:"log_level"`
	OutputDir string `json:"output_dir"`
}

// NewConfig creates a new configuration with default values
func NewConfig() Config {
	return &DefaultConfig{
		LogLevel:  "info",
		OutputDir: "output",
	}
}

// GetLogLevel returns the current log level
func (c *DefaultConfig) GetLogLevel() string {
	return c.LogLevel
}

// GetOutputDir returns the current output directory with ~ expanded
func (c *DefaultConfig) GetOutputDir() string {
	// Expand ~ to home directory
	if strings.HasPrefix(c.OutputDir, "~") {
		return filepath.Join(os.Getenv("HOME"), strings.TrimPrefix(c.OutputDir, "~"))
	}
	return c.OutputDir
}

// SetOutputDir sets the output directory
func (c *DefaultConfig) SetOutputDir(dir string) {
	c.OutputDir = dir
}

// Save saves the configuration to a file
func (c *DefaultConfig) Save() error {
	configDir := filepath.Join(os.Getenv("HOME"), ".phite")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	configFile, err := os.OpenFile(configPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer configFile.Close()

	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

// LoadConfig loads configuration from a file
func LoadConfig() (Config, error) {
	configDir := filepath.Join(os.Getenv("HOME"), ".phite")
	configPath := filepath.Join(configDir, "config.json")

	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		return NewConfig(), nil
	}

	configFile, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer configFile.Close()

	var config DefaultConfig
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return &config, nil
}
