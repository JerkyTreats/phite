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

// Exported configuration keys
const (
	LogLevelKey                        = "log_level"
	ReferenceGenomeBuildKey            = "reference_genome_build"
	PRSStatsCacheGCPProjectIDKey       = "prs_stats_cache.gcp_project_id"
	PRSStatsCacheDatasetIDKey          = "prs_stats_cache.dataset_id"
	PRSStatsCacheTableIDKey            = "prs_stats_cache.table_id"
	AlleleFreqSourceKey                = "allele_freq_source" // Parent key for the whole map
	AlleleFreqSourceTypeKey            = "allele_freq_source.type"
	AlleleFreqSourceGCPProjectIDKey    = "allele_freq_source.gcp_project_id"
	AlleleFreqSourceDatasetIDPatternKey = "allele_freq_source.dataset_id_pattern"
	AlleleFreqSourceTableIDPatternKey  = "allele_freq_source.table_id_pattern"
	AlleleFreqSourceAncestryMappingKey = "allele_freq_source.ancestry_mapping"
	PRSModelSourceTypeKey              = "prs_model_source.type"
	PRSModelSourcePathOrTableURIKey    = "prs_model_source.path_or_table_uri"
	PRSModelSNPIDColKey                = "prs_model_source.snp_id_column_name"
	PRSModelEffectAlleleColKey         = "prs_model_source.effect_allele_column_name"
	PRSModelOtherAlleleColKey          = "prs_model_source.other_allele_column_name"
	PRSModelWeightColKey               = "prs_model_source.weight_column_name"
	PRSModelChromosomeColKey           = "prs_model_source.chromosome_column_name"
	PRSModelPositionColKey             = "prs_model_source.position_column_name"
	PRSModelSourceModelIDColKey        = "prs_model_source.model_id_column_name" // Optional: Column name for the PRS model identifier (e.g., study_id)
	PRSModelSourceTableNameKey                  = "prs_model_source.table_name"                             // Optional: Name of the table in DuckDB (defaults to associations_clean)
	PRSModelSourceEffectAlleleFrequencyColKey = "prs_model_source.effect_allele_frequency_column_name"  // Optional
	PRSModelSourceBetaValueColKey             = "prs_model_source.beta_value_column_name"               // Optional
	PRSModelSourceBetaCILowerColKey           = "prs_model_source.beta_ci_lower_column_name"            // Optional
	PRSModelSourceBetaCIUpperColKey           = "prs_model_source.beta_ci_upper_column_name"            // Optional
	PRSModelSourceOddsRatioColKey             = "prs_model_source.odds_ratio_column_name"               // Optional
	PRSModelSourceORCILowerColKey             = "prs_model_source.or_ci_lower_column_name"              // Optional
	PRSModelSourceORCIUpperColKey             = "prs_model_source.or_ci_upper_column_name"              // Optional
	PRSModelSourceVariantIDColKey             = "prs_model_source.variant_id_column_name"               // Optional
	PRSModelSourceRSIDColKey                  = "prs_model_source.rsid_column_name"                     // Optional
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
	// 1. Required keys registered by other packages
	requiredKeysMutex.Lock()
	keysToCheck := make([]string, len(requiredKeys))
	copy(keysToCheck, requiredKeys)
	requiredKeysMutex.Unlock()
	for _, key := range keysToCheck {
		if !HasKey(key) {
			missingKeys = append(missingKeys, key)
		}
	}

	// 2. Enforce reference_genome_build == GRCh38 (case-insensitive)
	build := GetString(ReferenceGenomeBuildKey)
	if strings.TrimSpace(build) == "" || strings.ToUpper(build) != "GRCH38" {
		return fmt.Errorf("%s must be 'GRCh38' (case-insensitive), got '%s'", ReferenceGenomeBuildKey, build)
	}

	// 3. Validate presence of required PRS reference config keys
	// 3a. prs_stats_cache
	if !config.IsSet(PRSStatsCacheGCPProjectIDKey) {
		missingKeys = append(missingKeys, PRSStatsCacheGCPProjectIDKey)
	}
	if !config.IsSet(PRSStatsCacheDatasetIDKey) {
		missingKeys = append(missingKeys, PRSStatsCacheDatasetIDKey)
	}
	if !config.IsSet(PRSStatsCacheTableIDKey) {
		missingKeys = append(missingKeys, PRSStatsCacheTableIDKey)
	}
	// 3b. allele_freq_source
	if !config.IsSet(AlleleFreqSourceTypeKey) {
		missingKeys = append(missingKeys, AlleleFreqSourceTypeKey)
	}
	if !config.IsSet(AlleleFreqSourceGCPProjectIDKey) {
		missingKeys = append(missingKeys, AlleleFreqSourceGCPProjectIDKey)
	}
	if !config.IsSet(AlleleFreqSourceDatasetIDPatternKey) {
		missingKeys = append(missingKeys, AlleleFreqSourceDatasetIDPatternKey)
	}
	if !config.IsSet(AlleleFreqSourceTableIDPatternKey) {
		missingKeys = append(missingKeys, AlleleFreqSourceTableIDPatternKey)
	}
	if !config.IsSet(AlleleFreqSourceAncestryMappingKey) {
		missingKeys = append(missingKeys, AlleleFreqSourceAncestryMappingKey)
	} else {
		// Check if it's actually a map
		val := config.Get(AlleleFreqSourceAncestryMappingKey)
		switch val.(type) {
		case map[string]interface{}, map[string]string:
			// Correct type
		default:
			missingKeys = append(missingKeys, fmt.Sprintf("%s must be a map[string]string", AlleleFreqSourceAncestryMappingKey))
		}
	}
	// 3c. prs_model_source
	if !config.IsSet(PRSModelSourceTypeKey) {
		missingKeys = append(missingKeys, PRSModelSourceTypeKey)
	}
	if !config.IsSet(PRSModelSourcePathOrTableURIKey) {
		missingKeys = append(missingKeys, PRSModelSourcePathOrTableURIKey)
	}
	if !config.IsSet(PRSModelSNPIDColKey) {
		missingKeys = append(missingKeys, PRSModelSNPIDColKey)
	}
	if !config.IsSet(PRSModelEffectAlleleColKey) {
		missingKeys = append(missingKeys, PRSModelEffectAlleleColKey)
	}
	// OtherAlleleCol is optional for some PRS model formats, but good to have if available.
	// We won't make it strictly required by default in global config validation,
	// but the PRSReferenceDataSource might enforce it based on model type or specific needs.
	// if !config.IsSet(PRSModelOtherAlleleColKey) {
	// 	missingKeys = append(missingKeys, PRSModelOtherAlleleColKey)
	// }
	if !config.IsSet(PRSModelWeightColKey) {
		missingKeys = append(missingKeys, PRSModelWeightColKey)
	}
	if !config.IsSet(PRSModelChromosomeColKey) {
		missingKeys = append(missingKeys, PRSModelChromosomeColKey)
	}
	if !config.IsSet(PRSModelPositionColKey) {
		missingKeys = append(missingKeys, PRSModelPositionColKey)
	}

	if len(missingKeys) > 0 {
		return fmt.Errorf("missing required config keys: %v", missingKeys)
	}

	// 4. Validate log_level
	level := strings.ToUpper(GetString(LogLevelKey))
	switch level {
	case "TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL", "PANIC":
		// Valid log level
	default:
		return fmt.Errorf("invalid log_level '%s' (case-insensitive), must be one of: TRACE, DEBUG, INFO, WARN, ERROR, FATAL, PANIC", GetString(LogLevelKey))
	}
	return nil
}
