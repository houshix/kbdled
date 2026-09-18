package main

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
)

// struct input_event layout on Linux (x86/ARM, 64-bit):
//
//	struct timeval time (tv_sec int64 + tv_usec int64) = 16 bytes
//	__u16 type, __u16 code, __s32 value                = 8 bytes
//
// Total = 24 bytes. Fields are read at fixed byte offsets instead of via a
// Go struct, so this doesn't depend on how Go would pad/align the struct.
const (
	evKey         = 1
	evLED         = 0x11 // reported whenever the kernel changes any LED on this device
	inputEventLen = 24
)

var (
	ErrNoInputDevices = errors.New("no input devices found")
	ErrKeyTimeout     = errors.New("no key press detected before timeout")
)

type keyPress struct {
	device  string
	keycode uint16
}

// watchDeviceForKeypress reads one /dev/input/eventX until it sees a key
// press (value == 1) with a plausible keyboard code (code < 256, excluding
// the BTN_* range used by mice/joysticks), then sends it on out.
//
// Used only during "install"/"remap": it's a blocking read, so devices that
// lose the race stay blocked in I/O until the (short-lived) process exits -
// acceptable for a one-shot CLI command.
func watchDeviceForKeypress(ctx context.Context, path string, out chan<- keyPress) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	buf := make([]byte, inputEventLen)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if _, err := io.ReadFull(f, buf); err != nil {
			return
		}
		evType := binary.LittleEndian.Uint16(buf[16:18])
		code := binary.LittleEndian.Uint16(buf[18:20])
		value := int32(binary.LittleEndian.Uint32(buf[20:24]))

		if evType == evKey && value == 1 && code < 256 {
			select {
			case out <- keyPress{device: path, keycode: code}:
			case <-ctx.Done():
			}
			return
		}
	}
}

// captureKeypress listens on every /dev/input/eventN at once and returns the
// first key press detected, or a sentinel error.
func captureKeypress(timeout time.Duration) (keyPress, error) {
	devices, _ := filepath.Glob("/dev/input/event*")
	if len(devices) == 0 {
		return keyPress{}, ErrNoInputDevices
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	out := make(chan keyPress, 1)
	for _, dev := range devices {
		go watchDeviceForKeypress(ctx, dev, out)
	}

	select {
	case kp := <-out:
		return kp, nil
	case <-ctx.Done():
		return keyPress{}, ErrKeyTimeout
	}
}

// daemonEvent distinguishes what watchDaemonEvents saw on the wire.
type daemonEvent int

const (
	// evToggle: the configured key was pressed - flip the LED.
	evToggle daemonEvent = iota
	// evLEDChanged: the kernel changed some LED on this device on its
	// own (this is what Caps/Num Lock resyncing all three lock LEDs
	// together looks like from here) - check ours still matches what we
	// want and fix it if not.
	evLEDChanged
)

// watchDaemonEvents is the daemon's continuous reader: it keeps reading the
// already-connected device forever and reports every press of the
// configured keycode, plus every EV_LED event (which is how we notice the
// kernel clobbering our LED when Caps/Num Lock changes - there's no sysfs
// switch to stop that resync, so reacting to it here is the actual fix).
// It closes the channel if the device disappears (e.g. a USB keyboard
// unplugged); the caller is expected to reconnect rather than treat this
// as fatal.
func watchDaemonEvents(ctx context.Context, f *os.File, keycode uint16, out chan<- daemonEvent) {
	defer close(out)
	buf := make([]byte, inputEventLen)
	for {
		if _, err := io.ReadFull(f, buf); err != nil {
			return
		}
		evType := binary.LittleEndian.Uint16(buf[16:18])
		code := binary.LittleEndian.Uint16(buf[18:20])
		value := int32(binary.LittleEndian.Uint32(buf[20:24]))

		var ev daemonEvent
		switch {
		case evType == evKey && code == keycode && value == 1:
			ev = evToggle
		case evType == evLED:
			ev = evLEDChanged
		default:
			continue
		}

		select {
		case out <- ev:
		case <-ctx.Done():
			return
		}
	}
}

// stablePath resolves a /dev/input/eventN to a path that survives reboots
// (by-id first, by-path as a fallback). Without this, the eventN number can
// change and the daemon loses track of the configured device.
func stablePath(eventPath string) string {
	for _, dir := range []string{"/dev/input/by-id", "/dev/input/by-path"} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			link := filepath.Join(dir, e.Name())
			target, err := filepath.EvalSymlinks(link)
			if err != nil {
				continue
			}
			if target == eventPath {
				return link
			}
		}
	}
	warn(t("no_stable_path", eventPath))
	return eventPath
}
