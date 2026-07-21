// internal/config/config.go

package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"server"`

	Database struct {
		BasePath       string `json:"base_path"`
		WalMode        bool   `json:"wal_mode"`
		TimeoutSeconds int    `json:"timeout_seconds"`
	} `json:"database"`

	Session struct {
		SessionIDLength   int `json:"session_id_length"`
		SessionCodeLength int `json:"session_code_length"`
	} `json:"session"`

	Security struct {
		MaxNameLength int    `json:"max_name_length"`
		MaxRollLength int    `json:"max_roll_length"`
		RateLimit     string `json:"rate_limit"`
	} `json:"security"`

	Killswitch struct {
		InactivityTimeoutMinutes int `json:"inactivity_timeout_minutes"`
	} `json:"killswitch"`

	Export struct {
		BackupPath string `json:"backup_path"`
	} `json:"export"`
}

func LoadConfig(path string) (*Config, error) {
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		path = envPath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := Config{}
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("parsing config json: %w", err)
	}

	return &cfg, nil
}

func (c *Config) DefaultDB() string {
	return c.Database.BasePath + "default.db"
}
