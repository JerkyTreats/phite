package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// GarminConfig holds all config under the `garmin` object in ~/.phite/config.json
type GarminConfig struct {
	// Add Garmin-specific config fields here, e.g.:
	UserWeightKg   float64 `json:"user_weight_kg"`
	UserSex        string  `json:"user_sex"`
	UserAge        int     `json:"user_age"`
	SweatRateLph   float64 `json:"sweat_rate_lph"`
	// Add more as needed
}

type configFile struct {
	Garmin GarminConfig `json:"garmin"`
}

// LoadGarminConfig loads the garmin config from ~/.phite/config.json
func LoadGarminConfig() (*GarminConfig, error) {
	configPath := filepath.Join(os.Getenv("HOME"), ".phite", "config.json")
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var cfg configFile
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return &cfg.Garmin, nil
}
