package main

import (
	"os"
	"path/filepath"
	"strings"
)

func scrollLockLEDs() []string {
	matches, _ := filepath.Glob("/sys/class/leds/*::scrolllock")
	return matches
}

// setLED writes 0 or 1 to the brightness file of every scroll-lock LED.
func setLED(value int) {
	v := "0"
	if value != 0 {
		v = "1"
	}
	for _, led := range scrollLockLEDs() {
		_ = os.WriteFile(filepath.Join(led, "brightness"), []byte(v), 0644)
	}
}

// readCurrentTrigger returns the active trigger name from the LED's
// sysfs "trigger" file, e.g. "none [kbd-scrollock] rc-feedback" -> "kbd-scrollock".
func readCurrentTrigger(ledDir string) string {
	data, err := os.ReadFile(filepath.Join(ledDir, "trigger"))
	if err != nil {
		return ""
	}
	for _, tok := range strings.Fields(string(data)) {
		if strings.HasPrefix(tok, "[") && strings.HasSuffix(tok, "]") {
			return strings.Trim(tok, "[]")
		}
	}
	return ""
}

// disableTriggersAndRemember sets every scroll-lock LED's trigger to
// "none" and records the original (once) on cfg for uninstall to restore.
// sysfs trigger state resets on reboot, so this runs on every startup;
// re-running when already "none" is a no-op.
func disableTriggersAndRemember(cfg *Config) {
	if cfg.OriginalTriggers == nil {
		cfg.OriginalTriggers = map[string]string{}
	}
	changed := false
	for _, led := range scrollLockLEDs() {
		current := readCurrentTrigger(led)
		if current != "" && current != "none" {
			if _, known := cfg.OriginalTriggers[led]; !known {
				cfg.OriginalTriggers[led] = current
				changed = true
			}
		}
		_ = os.WriteFile(filepath.Join(led, "trigger"), []byte("none"), 0644)
	}
	if changed {
		_ = saveConfig(cfg)
	}
}

// reassertLEDIfNeeded forces the LED back to 1 if the kernel turned it
// off. Caps/Num Lock changes trigger a kernel-level resync of all lock
// LEDs (input-leds) that "trigger" can't stop, so this is the actual fix
// - called on EV_LED events and by a backup timer.
func reassertLEDIfNeeded() {
	for _, led := range scrollLockLEDs() {
		data, err := os.ReadFile(filepath.Join(led, "brightness"))
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(data)) == "0" {
			_ = os.WriteFile(filepath.Join(led, "brightness"), []byte("1"), 0644)
		}
	}
}

// restoreTriggers puts back each LED's original trigger, so uninstall
// doesn't leave keyboard LED behavior permanently disabled.
func restoreTriggers(cfg *Config) {
	for led, trig := range cfg.OriginalTriggers {
		_ = os.WriteFile(filepath.Join(led, "trigger"), []byte(trig), 0644)
	}
}
