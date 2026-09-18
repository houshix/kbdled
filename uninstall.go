package main

import (
	"fmt"
	"os"
)

func cmdUninstall() {
	requireRoot()

	// Load first: need OriginalTriggers to restore LED behavior.
	cfg, err := loadConfig()

	if err := run("systemctl", "disable", "--now", serviceName); err != nil {
		warn("systemctl disable failed: " + err.Error())
	}

	setLED(0) // don't leave the LED lit on the way out
	if err == nil {
		restoreTriggers(cfg)
	}

	os.Remove(binaryPath)
	os.RemoveAll("/etc/kbdled")
	os.Remove(unitPath)
	os.Remove(statePath)

	if err := run("systemctl", "daemon-reload"); err != nil {
		warn("systemctl daemon-reload failed: " + err.Error())
	}

	fmt.Println("Removed: service stopped, LED reset, all files deleted.")
}
