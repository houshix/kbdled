package main

import (
	"os"
	"strings"
)

// /run is tmpfs and is cleared on reboot, like /tmp - but it isn't
// world-writable, so another local user can't tamper with your LED state.
const statePath = "/run/kbdled.state"

func loadState() int {
	data, err := os.ReadFile(statePath)
	if err != nil {
		return 0
	}
	if strings.TrimSpace(string(data)) == "1" {
		return 1
	}
	return 0
}

func saveState(state int) {
	v := "0"
	if state != 0 {
		v = "1"
	}
	_ = os.WriteFile(statePath, []byte(v), 0644)
}
