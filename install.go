package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const captureTimeout = 20 * time.Second

func cmdInstall() {
	requireRoot()

	fmt.Printf("Press and hold the key combination you want to use, then release (%ds window).\n", int(captureTimeout.Seconds()))
	fmt.Println("Note: Fn is usually handled by the keyboard's own firmware, not the OS -")
	fmt.Println("a combination that includes it likely won't be seen here.")

	device, keycodes, err := captureKeyCombo(captureTimeout)
	if err != nil {
		reportCaptureError(err)
		os.Exit(1)
	}

	stable := stablePath(device)
	fmt.Println("Captured:", comboName(keycodes), "on", stable)

	if len(scrollLockLEDs()) == 0 {
		warn("warning: no matching LED found under /sys/class/leds (looking for *::scrolllock); the key will be captured, but nothing will light up until a matching LED exists")
	}

	cfg := &Config{Device: stable, Keycodes: keycodes}
	if err := saveConfig(cfg); err != nil {
		fatal("saving config", err)
	}

	if err := installBinary(); err != nil {
		fatal("installing binary", err)
	}
	if err := writeUnit(); err != nil {
		fatal("writing systemd unit", err)
	}

	if err := run("systemctl", "daemon-reload"); err != nil {
		warn("systemctl daemon-reload failed: " + err.Error())
	}
	if err := run("systemctl", "enable", "--now", serviceName); err != nil {
		warn("systemctl enable failed: " + err.Error())
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Installation complete. The key now works even on the login screen (SDDM/GDM/LightDM).")
	fmt.Println("If this combination is already bound to a shortcut in your desktop environment,")
	fmt.Println("remove that binding - this program now captures it directly, and both would fire together.")
}

// comboName joins keycode names, e.g. "Left Ctrl + Scroll Lock".
func comboName(codes []uint16) string {
	names := make([]string, len(codes))
	for i, c := range codes {
		names[i] = keyName(c)
	}
	return strings.Join(names, " + ")
}

func reportCaptureError(err error) {
	switch {
	case errors.Is(err, ErrNoInputDevices):
		warn("No input device found under /dev/input")
	case errors.Is(err, ErrKeyTimeout):
		warn("No key press detected within " + captureTimeout.String())
	default:
		warn(err.Error())
	}
}

// installBinary copies the running executable to /usr/local/bin. Writes
// to a temp file and renames into place (atomic, avoids ETXTBSY on a
// running binary) rather than truncating the target directly.
func installBinary() error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	if resolved, err := filepath.EvalSymlinks(self); err == nil {
		self = resolved
	}

	if same, _ := sameFile(self, binaryPath); same {
		return nil
	}

	src, err := os.Open(self)
	if err != nil {
		return err
	}
	defer src.Close()

	tmp := binaryPath + ".tmp"
	dst, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		os.Remove(tmp)
		return err
	}
	if err := dst.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, binaryPath)
}

func sameFile(a, b string) (bool, error) {
	fa, err := os.Stat(a)
	if err != nil {
		return false, err
	}
	fb, err := os.Stat(b)
	if err != nil {
		return false, nil // target missing
	}
	return os.SameFile(fa, fb), nil
}
