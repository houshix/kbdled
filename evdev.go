package main

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

// Linux input_event layout (x86/ARM, 64-bit): timeval (16 bytes) + type,
// code, value (8 bytes) = 24 bytes total. Read by offset, not a Go struct,
// to avoid depending on Go's own padding rules.
const (
	evKey         = 1
	evLED         = 0x11 // LED state change
	inputEventLen = 24
)

var (
	ErrNoInputDevices = errors.New("no input devices found")
	ErrKeyTimeout     = errors.New("no key press detected before timeout")
)

type comboResult struct {
	device   string
	keycodes []uint16
}

// captureKeyCombo waits for a key chord on whichever /dev/input/eventN
// reports the first plausible key press (code < 256, excludes mouse/
// joystick BTN_* codes). That device then tracks all further presses
// until the first release, which ends capture and returns the combo.
func captureKeyCombo(timeout time.Duration) (device string, keycodes []uint16, err error) {
	devices, _ := filepath.Glob("/dev/input/event*")
	if len(devices) == 0 {
		return "", nil, ErrNoInputDevices
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var won int32
	resultCh := make(chan comboResult, 1)
	for _, dev := range devices {
		go captureOnDevice(ctx, dev, &won, resultCh)
	}

	select {
	case r := <-resultCh:
		return r.device, r.keycodes, nil
	case <-ctx.Done():
		return "", nil, ErrKeyTimeout
	}
}

// captureOnDevice reads one device. The first goroutine to see a
// qualifying press claims the race via `won` and tracks the combo;
// losers exit immediately. Non-matching devices (mice, etc.) block until
// process exit, which is fine for a short-lived CLI step.
func captureOnDevice(ctx context.Context, path string, won *int32, out chan<- comboResult) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	buf := make([]byte, inputEventLen)
	var pressed []uint16
	claimed := false

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

		if evType != evKey || code >= 256 {
			continue
		}

		if !claimed {
			if value != 1 {
				continue
			}
			if !atomic.CompareAndSwapInt32(won, 0, 1) {
				return // lost the race
			}
			claimed = true
			pressed = append(pressed, code)
			continue
		}

		switch value {
		case 1: // another key joined the combo
			if !containsCode(pressed, code) {
				pressed = append(pressed, code)
			}
		case 0: // any release ends the combo
			select {
			case out <- comboResult{device: path, keycodes: pressed}:
			case <-ctx.Done():
			}
			return
		}
	}
}

func containsCode(list []uint16, c uint16) bool {
	for _, v := range list {
		if v == c {
			return true
		}
	}
	return false
}

// daemonEvent distinguishes what watchDaemonEvents saw on the wire.
type daemonEvent int

const (
	evToggle     daemonEvent = iota // combo fully held: flip the LED
	evLEDChanged                    // kernel changed a LED; re-check ours
)

// watchDaemonEvents tracks the configured keycodes and reports evToggle on
// the rising edge of "all held" (no retrigger while held). It also reports
// every EV_LED event, since that's the only way to catch the kernel
// clobbering our LED on Caps/Num Lock. Closes the channel on disconnect;
// the caller should reconnect, not treat it as fatal.
func watchDaemonEvents(ctx context.Context, f *os.File, keycodes []uint16, out chan<- daemonEvent) {
	defer close(out)
	buf := make([]byte, inputEventLen)
	pressed := make(map[uint16]bool, len(keycodes))
	wasAllHeld := false

	allHeld := func() bool {
		for _, k := range keycodes {
			if !pressed[k] {
				return false
			}
		}
		return true
	}

	for {
		if _, err := io.ReadFull(f, buf); err != nil {
			return
		}
		evType := binary.LittleEndian.Uint16(buf[16:18])
		code := binary.LittleEndian.Uint16(buf[18:20])
		value := int32(binary.LittleEndian.Uint32(buf[20:24]))

		switch {
		case evType == evKey && code < 256:
			switch value {
			case 1:
				pressed[code] = true
			case 0:
				pressed[code] = false
			}
			nowAllHeld := allHeld()
			if nowAllHeld && !wasAllHeld {
				select {
				case out <- evToggle:
				case <-ctx.Done():
					return
				}
			}
			wasAllHeld = nowAllHeld

		case evType == evLED:
			select {
			case out <- evLEDChanged:
			case <-ctx.Done():
				return
			}
		}
	}
}

// stablePath resolves eventN to a reboot-stable path (by-id, then
// by-path); eventN numbering can change across boots.
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
	warn("warning: no stable path (by-id/by-path) found for " + eventPath + "; the device name may change across reboots")
	return eventPath
}
