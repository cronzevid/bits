package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds all configuration settings
type Config struct {
	Telegram struct {
		Token string `json:"token"`
	} `json:"telegram"`
	Bitstamp struct {
		Key    string `json:"key"`
		Secret string `json:"secret"`
		Uid    string `json:"uid"`
	} `json:"bitstamp"`
	Notifications struct {
		HourlyThreshold float64 `json:"hourly_threshold"`
		DailyThreshold  float64 `json:"daily_threshold"`
		WeeklyThreshold float64 `json:"weekly_threshold"`
		RetryCount      int     `json:"retry_count"`
		RetryDelay      int     `json:"retry_delay"`
	} `json:"notifications"`
}

// BitstampConfig returns the Bitstamp-specific configuration
func (c *Config) BitstampConfig() BitstampConfig {
	return BitstampConfig{
		Key:    c.Bitstamp.Key,
		Secret: c.Bitstamp.Secret,
		Id:     c.Bitstamp.Uid,
	}
}

// BitstampConfig is used for Bitstamp API authentication
type BitstampConfig struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
	Id     string `json:"id"`
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	// If no path provided, use default
	if configPath == "" {
		configPath = "config.json"
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, err
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Parse config
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	// If no path provided, use default
	if configPath == "" {
		configPath = "config.json"
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Marshal config
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Write file
	return os.WriteFile(configPath, data, 0644)
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	config := &Config{}

	// Set default values
	config.Notifications.HourlyThreshold = 3.0
	config.Notifications.DailyThreshold = 5.0
	config.Notifications.WeeklyThreshold = 10.0
	config.Notifications.RetryCount = 3
	config.Notifications.RetryDelay = 5

	return config
}
