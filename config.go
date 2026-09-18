package main

import (
	"encoding/json"
	"os"
)

const configPath = "/etc/kbdled/config.json"

// Config holds the input device (stable path) and keycode chosen during
// "install" or "remap".
type Config struct {
	Device  string `json:"device"`
	Keycode uint16 `json:"keycode"`
	// Lang remembers the wizard's last chosen language, so re-running
	// install/remap/uninstall pre-selects it instead of defaulting to English.
	Lang string `json:"lang,omitempty"`
	// OriginalTriggers maps an LED's sysfs directory to the trigger name
	// the kernel had assigned before we switched it to "none", so
	// uninstall can restore normal keyboard LED behavior.
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
