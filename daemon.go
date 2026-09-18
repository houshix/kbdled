package main

import (
	"context"
	"fmt"
	"os"
	"time"
)

// reassertInterval is a low-frequency safety net on top of the reactive
// EV_LED handling, in case some driver/keyboard combination doesn't
// surface a clean EV_LED event for a resync we need to undo.
const reassertInterval = 2 * time.Second

// reconnectDelay is how long to wait between attempts to (re)open the
// configured device, both when it's missing at startup and when it goes
// away while running.
const reconnectDelay = 2 * time.Second

func cmdDaemon() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "no configuration found, run 'sudo kbdled install' first:", err)
		os.Exit(1)
	}

	// This loop normally never exits: a missing or disconnected keyboard
	// is handled by waiting and reconnecting in-process, rather than by
	// exiting and relying on systemd to restart the whole service. That
	// matters because a keyboard that's still unplugged a few seconds
	// after boot would otherwise make the daemon fail repeatedly in a
	// short window and trip systemd's restart rate limit, requiring a
	// fresh reboot (with the keyboard already connected) to recover.
	for {
		runOnce(cfg)
		time.Sleep(reconnectDelay)
	}
}

// runOnce waits for the configured device, applies the desired LED state,
// and services it until it disconnects.
func runOnce(cfg *Config) {
	f := waitForDevice(cfg.Device)
	defer f.Close()

	if len(scrollLockLEDs()) == 0 {
		fmt.Fprintln(os.Stderr, "warning: no matching LED found under /sys/class/leds")
	}

	// Every (re)connection gets a fresh sysfs LED node (unplugging a USB
	// keyboard destroys the old one), so this needs to run again here,
	// not just once at process startup.
	disableTriggersAndRemember(cfg)

	state := loadState()
	setLED(state)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events := make(chan daemonEvent)
	go watchDaemonEvents(ctx, f, cfg.Keycode, events)

	ticker := time.NewTicker(reassertInterval)
	defer ticker.Stop()

	for {
		select {
		case ev, ok := <-events:
			if !ok {
				fmt.Fprintln(os.Stderr, "input device disconnected, waiting to reconnect")
				return
			}
			switch ev {
			case evToggle:
				state = 1 - state
				saveState(state)
				setLED(state)
			case evLEDChanged:
				if state == 1 {
					reassertLEDIfNeeded()
				}
			}
		case <-ticker.C:
			if state == 1 {
				reassertLEDIfNeeded()
			}
		}
	}
}

// waitForDevice retries opening the configured input device until it
// appears - covers both a keyboard that's still unplugged when the daemon
// starts and one that gets unplugged and later reconnected.
func waitForDevice(path string) *os.File {
	for {
		if f, err := os.Open(path); err == nil {
			return f
		}
		time.Sleep(reconnectDelay)
	}
}
