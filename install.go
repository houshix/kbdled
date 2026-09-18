package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const captureTimeout = 20 * time.Second

func cmdInstall() {
	requireRoot()
	pickLanguage()

	fmt.Println(t("press_key", int(captureTimeout.Seconds())))
	kp, err := captureKeypress(captureTimeout)
	if err != nil {
		reportCaptureError(err)
		os.Exit(1)
	}

	device := stablePath(kp.device)
	fmt.Println(t("key_captured", kp.keycode, device))

	if len(scrollLockLEDs()) == 0 {
		warn(t("no_led_found"))
	}

	cfg := &Config{Device: device, Keycode: kp.keycode, Lang: lang}
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
	fmt.Println(t("install_done"))
	fmt.Println(t("shortcut_note"))
}

func reportCaptureError(err error) {
	switch {
	case errors.Is(err, ErrNoInputDevices):
		warn(t("no_input_device"))
	case errors.Is(err, ErrKeyTimeout):
		warn(t("no_key_detected", captureTimeout))
	default:
		warn(err.Error())
	}
}

// installBinary copies the running executable to /usr/local/bin, unless
// it's already installed there. It writes to a temp file and renames it
// into place rather than truncating the target in place: renaming swaps
// the directory entry atomically without touching the inode a running
// process has open, avoiding ETXTBSY if kbdled is somehow re-run from the
// very path it's about to install to under a different name.
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
		return false, nil // target doesn't exist yet, so it can't be the same file
	}
	return os.SameFile(fa, fb), nil
}
