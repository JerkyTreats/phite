// Package config provides centralized, extensible configuration loading for PHITE using spf13/viper.
// All config access must go through this package.
package config

import (
	"os" // Added for ToUpper
	"sync"

	"github.com/spf13/viper"
)

// Shared infrastructure configuration keys - used across multiple domains
const (
	// Logging (core infrastructure)
	LogLevelKey = "logging.level"

	// GCP Project Infrastructure - addresses duplication across domains
	GCPDataProjectKey    = "gcp.data_project"    // Where data lives (e.g., bigquery-public-data)
	GCPBillingProjectKey = "gcp.billing_project" // User's project for query billing (required for public datasets)
	GCPCacheProjectKey   = "gcp.cache_project"   // Where user stores private cache tables

	// BigQuery Dataset Infrastructure - shared across repositories
	BigQueryGnomadDatasetKey = "bigquery.gnomad_dataset" // gnomAD dataset name
	BigQueryCacheDatasetKey  = "bigquery.cache_dataset"  // Cache dataset name

	// Table Infrastructure - shared table references
	TableCacheTableKey      = "tables.cache_table"       // Reference stats cache table
	TableModelTableKey      = "tables.model_table"       // PRS model table
	TableAlleleFreqTableKey = "tables.allele_freq_table" // Allele frequency table
)

// NOTE: Domain-specific configuration constants are defined in their respective packages:
// - ancestry.PopulationKey, ancestry.GenderKey -> internal/ancestry/config.go
// - invariance.EnableValidationKey, invariance.StrictModeKey -> internal/invariance/invariance.go
// - cache.BatchSizeKey -> internal/reference/cache/cache.go
// This maintains domain ownership while eliminating infrastructure duplication.

var (
	config            *viper.Viper
	configOnce        sync.Once
	configPath        string
	requiredKeys      []string
	requiredKeysMutex sync.Mutex
	// Replace global variable with a slice to track missing required keys
	MissingKeys []string
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
	v.SetDefault(LogLevelKey, "INFO")
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

// GetStringMapString returns a map[string]string config value.
func GetStringMapString(key string) map[string]string {
	_ = initConfig()
	if config == nil {
		return make(map[string]string) // Return empty map if config not loaded
	}
	return config.GetStringMapString(key)
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
	// Check if the key is present in the config
	if !HasKey(key) {
		MissingKeys = append(MissingKeys, key)
	}
}

// HasKey returns true if the config has the key.
func HasKey(key string) bool {
	_ = initConfig()
	if config == nil {
		return false
	}
	return config.IsSet(key)
}

// Set sets a configuration value for testing purposes only.
func Set(key string, value interface{}) {
	_ = initConfig()
	if config != nil {
		config.Set(key, value)
	}
}
