package main

import (
	"fmt"
	"os"
	"time"
)

const escSequenceTimeout = 150 * time.Millisecond

type menuItem struct {
	label    string
	disabled bool
	note     string // shown next to a disabled item
}

// selectMenu is an arrow-key list, confirmed with Enter, Space, or numpad
// Enter. Disabled items are dimmed and skipped when navigating. Falls
// back to a numbered prompt if stdin isn't a terminal.
//
// One goroutine owns all stdin reads and feeds a channel; escape sequences
// are disambiguated by timing out that channel receive, since Go's runtime
// poller makes stdin non-blocking and the kernel's VTIME is ignored as a
// result. This way no byte is ever dropped on a timeout.
func selectMenu(items []menuItem) int {
	fd := os.Stdin.Fd()
	orig, err := enableRawMode(fd)
	if err != nil {
		return numberedFallback(items)
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

	selected := firstEnabled(items)
	printMenu(items, selected)

	for b := range bytesCh {
		switch b {
		case 27: // ESC - lone Escape cancels; also starts "ESC [ A/B" and "ESC O M"
			b2, ok := readByteWithTimeout(bytesCh, escSequenceTimeout)
			if !ok {
				restoreMode(fd, orig) // lone Escape: nothing followed in time
				os.Exit(130)
			}
			switch b2 {
			case '[':
				b3, ok := readByteWithTimeout(bytesCh, escSequenceTimeout)
				if !ok {
					continue
				}
				switch b3 {
				case 'A': // Up
					selected = moveSelection(items, selected, -1)
					redrawMenu(items, selected)
				case 'B': // Down
					selected = moveSelection(items, selected, 1)
					redrawMenu(items, selected)
				}
			case 'O': // app keypad mode (numpad keys on some terminals)
				b3, ok := readByteWithTimeout(bytesCh, escSequenceTimeout)
				if !ok {
					continue
				}
				if b3 == 'M' { // numpad Enter
					fmt.Println()
					return selected
				}
			}
		case '\r', '\n', ' ': // Enter or Space confirms
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

func firstEnabled(items []menuItem) int {
	for i, it := range items {
		if !it.disabled {
			return i
		}
	}
	return 0
}

// moveSelection steps by delta, wrapping and skipping disabled items.
func moveSelection(items []menuItem, from, delta int) int {
	n := len(items)
	i := from
	for k := 0; k < n; k++ {
		i = (i + delta + n) % n
		if !items[i].disabled {
			return i
		}
	}
	return from
}

func printMenu(items []menuItem, selected int) {
	for i, it := range items {
		fmt.Print("\r\x1b[2K") // clear the line
		switch {
		case it.disabled:
			label := it.label
			if it.note != "" {
				label += "  (" + it.note + ")"
			}
			fmt.Printf("    \x1b[2m%s\x1b[0m\n", label) // dim
		case i == selected:
			fmt.Printf("  \x1b[36m> %s\x1b[0m\n", it.label)
		default:
			fmt.Printf("    %s\n", it.label)
		}
	}
}

func redrawMenu(items []menuItem, selected int) {
	fmt.Printf("\x1b[%dA", len(items)) // cursor back to top
	printMenu(items, selected)
}

func numberedFallback(items []menuItem) int {
	for i, it := range items {
		suffix := ""
		if it.disabled {
			suffix = "  (unavailable"
			if it.note != "" {
				suffix += " - " + it.note
			}
			suffix += ")"
		}
		fmt.Printf("  %d) %s%s\n", i+1, it.label, suffix)
	}
	fallback := firstEnabled(items)
	for {
		fmt.Print("> ")
		var choice int
		if _, err := fmt.Scanln(&choice); err != nil {
			return fallback
		}
		if choice < 1 || choice > len(items) {
			continue
		}
		if items[choice-1].disabled {
			fmt.Println("That option isn't available yet.")
			continue
		}
		return choice - 1
	}
}
