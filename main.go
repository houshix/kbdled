package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "install":
		cmdInstall()
	case "uninstall":
		cmdUninstall()
	case "remap":
		cmdRemap()
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
	fmt.Println(`kbdled - toggle a keyboard LED (scroll lock) with a dedicated key,
working even on the login screen (SDDM/GDM/LightDM), independent of any
graphical session.

Usage:
  sudo kbdled install     Interactive wizard: pick a key, install the service
  sudo kbdled remap       Change the key without a full reinstall
  sudo kbdled uninstall   Stop the service, reset the LED, remove all files
  kbdled daemon           Internal use - invoked by the systemd unit

The install/remap/uninstall wizards are interactive and support English,
Portuguese, Spanish, German, French and Chinese.`)
}
