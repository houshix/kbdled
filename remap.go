package main

import (
	"fmt"
	"os"
)

func cmdRemap() {
	requireRoot()

	cfg, err := loadConfig()
	if err != nil {
		warn("No configuration found. Run 'sudo kbdled install' first.")
		os.Exit(1)
	}

	fmt.Printf("Press and hold the new key combination you want to use, then release (%ds window).\n", int(captureTimeout.Seconds()))
	fmt.Println("Note: Fn is usually handled by the keyboard's own firmware, not the OS -")
	fmt.Println("a combination that includes it likely won't be seen here.")

	device, keycodes, err := captureKeyCombo(captureTimeout)
	if err != nil {
		reportCaptureError(err)
		os.Exit(1)
	}

	stable := stablePath(device)
	fmt.Println("Captured:", comboName(keycodes), "on", stable)

	// Update in place to keep OriginalTriggers for uninstall.
	cfg.Device = stable
	cfg.Keycodes = keycodes
	if err := saveConfig(cfg); err != nil {
		fatal("saving config", err)
	}

	fmt.Println("Key combination updated. Restarting the service...")
	if err := run("systemctl", "restart", serviceName); err != nil {
		warn("systemctl restart failed: " + err.Error())
	}
}
