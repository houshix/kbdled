package main

import (
	"fmt"
	"os"
)

func cmdRemap() {
	requireRoot()
	pickLanguage()

	cfg, err := loadConfig()
	if err != nil {
		warn(t("config_missing"))
		os.Exit(1)
	}

	fmt.Println(t("remap_prompt", int(captureTimeout.Seconds())))
	kp, err := captureKeypress(captureTimeout)
	if err != nil {
		reportCaptureError(err)
		os.Exit(1)
	}

	device := stablePath(kp.device)
	fmt.Println(t("key_captured", kp.keycode, device))

	// Update in place rather than building a fresh Config, so Lang and
	// OriginalTriggers (needed by uninstall later) aren't lost.
	cfg.Device = device
	cfg.Keycode = kp.keycode
	cfg.Lang = lang
	if err := saveConfig(cfg); err != nil {
		fatal("saving config", err)
	}

	fmt.Println(t("remap_done"))
	if err := run("systemctl", "restart", serviceName); err != nil {
		warn("systemctl restart failed: " + err.Error())
	}
}
