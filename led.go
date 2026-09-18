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

// readCurrentTrigger returns the LED trigger currently marked active in
// its sysfs "trigger" file, e.g. "none kbd-scrollock [rc-feedback]" -> "rc-feedback".
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

// disableTriggersAndRemember switches every scroll-lock LED's trigger to
// "none", so the kernel stops driving it on its own (it otherwise resyncs
// all three lock LEDs together whenever Caps/Num Lock state changes,
// clobbering whatever we last wrote to brightness). The first time it sees
// a real trigger for a given LED path, it records it on cfg so it can be
// restored on uninstall - sysfs LED state doesn't survive a reboot, so
// this needs to run once per daemon startup, and re-running it when the
// trigger is already "none" is a harmless no-op.
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

// reassertLEDIfNeeded forces the LED back to 1 if the kernel turned it off.
// This is the actual fix for Caps/Num Lock clobbering scroll lock's LED:
// that resync happens in a lower-level kernel path (input-leds, which
// re-syncs all three lock LEDs together on any of them changing) that
// "trigger" has no effect on, so the only reliable fix is to notice it
// happened and put our own value back. Called both reactively (on an
// EV_LED event from the device) and from a low-frequency safety-net timer.
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

// restoreTriggers puts back whatever trigger the kernel originally had on
// each LED, so uninstalling this program doesn't leave the system's normal
// keyboard LED behavior permanently disabled.
func restoreTriggers(cfg *Config) {
	for led, trig := range cfg.OriginalTriggers {
		_ = os.WriteFile(filepath.Join(led, "trigger"), []byte(trig), 0644)
	}
}
