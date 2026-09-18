package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		runWizard()
		return
	}

	switch os.Args[1] {
	case "install":
		cmdInstall()
	case "remap":
		cmdRemap()
	case "uninstall":
		cmdUninstall()
	case "daemon":
		cmdDaemon()
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`kbdled - toggle a keyboard LED (scroll lock) with a dedicated key or key
combination, working even on the login screen (SDDM/GDM/LightDM),
independent of any graphical session.

Usage:
  kbdled                   Interactive wizard (arrow keys; Space/Enter to confirm)
  sudo kbdled install      Set up the key combination and install the service
  sudo kbdled remap        Change the key combination without a full reinstall
  sudo kbdled uninstall    Stop the service, reset the LED, remove all files
  kbdled daemon            Internal use - invoked by the systemd unit`)
}
