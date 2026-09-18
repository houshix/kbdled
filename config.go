package main

import (
	"encoding/json"
	"os"
)

const configPath = "/etc/kbdled/config.json"

// Config is the persisted setup: input device, key combo, and each LED's
// original trigger (for uninstall to restore).
type Config struct {
	Device   string   `json:"device"`
	Keycodes []uint16 `json:"keycodes"`
	// OriginalTriggers: LED sysfs dir -> its trigger before we set "none".
	OriginalTriggers map[string]string `json:"original_triggers,omitempty"`
}

func loadConfig() (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func saveConfig(cfg *Config) error {
	if err := os.MkdirAll("/etc/kbdled", 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}
