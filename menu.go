package main

import (
	"fmt"
	"os"
	"time"
)

const escSequenceTimeout = 150 * time.Millisecond

// selectMenu renders options as an arrow-key selectable list (Up/Down +
// Enter) and returns the chosen index, starting with `initial` highlighted.
// Falls back to a plain numbered prompt when stdin isn't a terminal (e.g.
// piped input, non-interactive install).
//
// A single goroutine owns every read of stdin for the life of the menu and
// publishes each byte on a channel; disambiguating an escape sequence is
// done by timing out that channel receive (not by trying to rely on the
// terminal's own VTIME, which Go's runtime poller puts stdin in non-blocking
// mode and its own readiness wait ends up ignoring). Routing every byte
// through the one channel means a byte is never silently dropped even when
// a read "times out".
func selectMenu(options []string, initial int) int {
	fd := os.Stdin.Fd()
	orig, err := enableRawMode(fd)
	if err != nil {
		return numberedFallback(options)
	}
	cancel := restoreOnSignal(fd, orig)
	defer func() {
		cancel()
		restoreMode(fd, orig)
	}()

	bytesCh := make(chan byte)
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				close(bytesCh)
				return
			}
			bytesCh <- buf[0]
		}
	}()

	selected := initial
	if selected < 0 || selected >= len(options) {
		selected = 0
	}
	printMenu(options, selected)

	for b := range bytesCh {
		switch b {
		case 27: // ESC - either a lone Escape (cancel) or "ESC [ A/B" (arrow)
			b2, ok := readByteWithTimeout(bytesCh, escSequenceTimeout)
			if !ok {
				restoreMode(fd, orig) // lone Escape: nothing followed in time
				os.Exit(130)
			}
			if b2 != '[' {
				continue
			}
			b3, ok := readByteWithTimeout(bytesCh, escSequenceTimeout)
			if !ok {
				continue
			}
			switch b3 {
			case 'A':
				selected = (selected - 1 + len(options)) % len(options)
				redrawMenu(options, selected)
			case 'B':
				selected = (selected + 1) % len(options)
				redrawMenu(options, selected)
			}
		case '\r', '\n':
			fmt.Println()
			return selected
		case 3: // Ctrl+C
			restoreMode(fd, orig)
			os.Exit(130)
		}
	}
	return selected
}

func readByteWithTimeout(ch <-chan byte, timeout time.Duration) (byte, bool) {
	select {
	case b, ok := <-ch:
		return b, ok
	case <-time.After(timeout):
		return 0, false
	}
}

func printMenu(options []string, selected int) {
	for i, opt := range options {
		fmt.Print("\r\x1b[2K") // return to column 0, clear the line
		if i == selected {
			fmt.Printf("  \x1b[36m> %s\x1b[0m\n", opt)
		} else {
			fmt.Printf("    %s\n", opt)
		}
	}
}

func redrawMenu(options []string, selected int) {
	fmt.Printf("\x1b[%dA", len(options)) // cursor back to the top of the menu
	printMenu(options, selected)
}

func numberedFallback(options []string) int {
	for i, opt := range options {
		fmt.Printf("  %d) %s\n", i+1, opt)
	}
	fmt.Print("> ")
	var choice int
	fmt.Scanln(&choice)
	if choice < 1 || choice > len(options) {
		return 0
	}
	return choice - 1
}
