package main

import (
	"context"
	"fmt"
	"os"
	"time"
)

// reassertInterval is a backup poll in case an EV_LED resync is missed.
const reassertInterval = 2 * time.Second

// reconnectDelay is the wait between attempts to (re)open the device.
const reconnectDelay = 2 * time.Second

func cmdDaemon() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "no configuration found, run 'sudo kbdled install' first:", err)
		os.Exit(1)
	}
	if len(cfg.Keycodes) == 0 {
		fmt.Fprintln(os.Stderr, "no key combination configured (old config format?), run 'sudo kbdled install' again")
		os.Exit(1)
	}

	// Never exits: reconnects in-process instead of relying on systemd
	// restarts, which could hit the rate limit if unplugged at boot.
	for {
		runOnce(cfg)
		time.Sleep(reconnectDelay)
	}
}

// runOnce connects, applies the LED state, and serves events until the
// device disconnects.
func runOnce(cfg *Config) {
	f := waitForDevice(cfg.Device)
	defer f.Close()

	if len(scrollLockLEDs()) == 0 {
		fmt.Fprintln(os.Stderr, "warning: no matching LED found under /sys/class/leds")
	}

	// Reconnecting gets a fresh sysfs LED node, so redo this each time.
	disableTriggersAndRemember(cfg)

	state := loadState()
	setLED(state)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events := make(chan daemonEvent)
	go watchDaemonEvents(ctx, f, cfg.Keycodes, events)

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

// waitForDevice retries until the input device appears.
func waitForDevice(path string) *os.File {
	for {
		if f, err := os.Open(path); err == nil {
			return f
		}
		time.Sleep(reconnectDelay)
	}
}
