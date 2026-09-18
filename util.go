package main

import (
	"fmt"
	"os"
	"os/exec"
)

// requireRoot is checked before the language picker: no point rendering an
// interactive menu just to fail right after on a precondition.
func requireRoot() {
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "Error: this command must be run as root (sudo).")
		os.Exit(1)
	}
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func fatal(msg string, err error) {
	fmt.Fprintf(os.Stderr, "error %s: %v\n", msg, err)
	os.Exit(1)
}

func warn(msg string) {
	fmt.Fprintln(os.Stderr, msg)
}
