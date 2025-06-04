// Package config provides centralized, extensible configuration loading for PHITE using spf13/viper.
// All config access must go through this package.
package config

import (
	"fmt"
	"os"
	"strings" // Added for ToUpper
	"sync"

	"github.com/spf13/viper"
)

var (
	config            *viper.Viper
	configOnce        sync.Once
	configPath        string
	requiredKeys      []string
	requiredKeysMutex sync.Mutex
)

// resetConfig is for test use only; resets the singleton.
// ResetForTest resets the config singleton for test use only.
func ResetForTest() {
	config = nil
	configOnce = sync.Once{}
	configPath = ""
	requiredKeysMutex.Lock()
	requiredKeys = nil
	requiredKeysMutex.Unlock()
}

// SetConfigPath allows test code to override the config file path before first use.
func SetConfigPath(path string) {
	configPath = path
}

// loadConfig initializes viper and loads config from file and env.
func loadConfig() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("json")
	v.SetConfigName("config")
	v.AddConfigPath(os.ExpandEnv("$HOME/.phite"))
	if configPath != "" {
		v.SetConfigFile(configPath)
	}
	v.AutomaticEnv()
	v.SetDefault("log_level", "INFO")
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// File not found: return viper instance with defaults
			return v, nil
		}
		// For parse errors or other errors, log and return viper with defaults
		return v, nil
	}
	return v, nil
}

// initConfig ensures config is loaded once.
func initConfig() error {
	var err error
	configOnce.Do(func() {
		var c *viper.Viper
		c, err = loadConfig()
		if err == nil {
			config = c
		} else {
			config = nil
		}
	})
	return err
}

// Reload reloads the configuration from disk (for hot reload, optional).
func Reload() error {
	c, err := loadConfig()
	if err != nil {
		return err
	}
	config = c
	return nil
}

// GetString returns a string config value.
func GetString(key string) string {
	_ = initConfig()
	if config == nil {
		// Return reasonable default for string
		return ""
	}
	return config.GetString(key)
}

// GetInt returns an int config value.
func GetInt(key string) int {
	_ = initConfig()
	if config == nil {
		return 0
	}
	return config.GetInt(key)
}

// GetBool returns a bool config value.
func GetBool(key string) bool {
	_ = initConfig()
	if config == nil {
		return false
	}
	return config.GetBool(key)
}

// RegisterRequiredKey adds a key to the list of required configuration items.
// This should be called during the init() phase of packages that require specific configurations.
func RegisterRequiredKey(key string) {
	requiredKeysMutex.Lock()
	defer requiredKeysMutex.Unlock()
	// Avoid duplicates
	for _, k := range requiredKeys {
		if k == key {
			return
		}
	}
	requiredKeys = append(requiredKeys, key)
}

// HasKey returns true if the config has the key.
func HasKey(key string) bool {
	_ = initConfig()
	if config == nil {
		return false
	}
	return config.IsSet(key)
}

// Validate checks for required/invalid config values.
func Validate() error {
	_ = initConfig() // Ensure config is loaded
	if config == nil {
		return fmt.Errorf("config not initialized, cannot validate")
	}

	var missingKeys []string
	requiredKeysMutex.Lock()
	keysToCheck := make([]string, len(requiredKeys))
	copy(keysToCheck, requiredKeys)
	requiredKeysMutex.Unlock()

	for _, key := range keysToCheck {
		if !HasKey(key) {
			missingKeys = append(missingKeys, key)
		}
	}

	if len(missingKeys) > 0 {
		return fmt.Errorf("missing required config keys: %v", missingKeys)
	}

	// Example: ensure log_level is valid (can be extended for other specific validations)
	lvl := GetString("log_level")
	switch strings.ToUpper(lvl) { // Convert to uppercase for case-insensitive comparison
	case "DEBUG", "INFO", "ERROR", "WARN":
		// Valid log level
	default:
		return fmt.Errorf("invalid log_level: %s. Must be one of DEBUG, INFO, ERROR, WARN (case-insensitive)", lvl)
	}
	return nil
}
