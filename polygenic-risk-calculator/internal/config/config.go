// Package config provides centralized, extensible configuration loading for PHITE using spf13/viper.
// All config access must go through this package.
package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
)

var (
	config     *viper.Viper
	configOnce sync.Once
	configPath string
)

// resetConfig is for test use only; resets the singleton.
// ResetForTest resets the config singleton for test use only.
func ResetForTest() {
	config = nil
	configOnce = sync.Once{}
	configPath = ""
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
	// Example: ensure log_level is valid
	lvl := GetString("log_level")
	switch lvl {
	case "DEBUG", "INFO", "ERROR", "WARN":
		return nil
	default:
		return fmt.Errorf("invalid log_level: %s", lvl)
	}
}
