package main

import (
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

// Minimal raw-mode terminal handling for Linux (x86/ARM), just to read
// arrow-key escape sequences one byte at a time for the interactive menu.
// This intentionally avoids golang.org/x/term (and its transitive
// dependency on golang.org/x/sys) to keep the module dependency-free -
// the kernel's struct termios (used by the TCGETS/TCSETS ioctls) has had
// this exact layout on every mainstream Linux architecture (x86, x86_64,
// ARM, ARM64) for decades.
type termios struct {
	Iflag uint32
	Oflag uint32
	Cflag uint32
	Lflag uint32
	Line  uint8
	Cc    [19]uint8
}

const (
	tcgets = 0x5401
	tcsets = 0x5402
	icanon = 0x0002
	echo   = 0x0008
	vtime  = 5
	vmin   = 6
)

func ioctlPtr(fd uintptr, req uintptr, arg unsafe.Pointer) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, req, uintptr(arg))
	if errno != 0 {
		return errno
	}
	return nil
}

// enableRawMode disables canonical mode and echo so keys - including arrow
// key escape sequences - can be read one byte at a time instead of waiting
// for a full line. Returns the original settings so they can be restored.
func enableRawMode(fd uintptr) (*termios, error) {
	var orig termios
	if err := ioctlPtr(fd, tcgets, unsafe.Pointer(&orig)); err != nil {
		return nil, err
	}
	raw := orig
	raw.Lflag &^= icanon | echo
	raw.Cc[vmin] = 1
	raw.Cc[vtime] = 0
	if err := ioctlPtr(fd, tcsets, unsafe.Pointer(&raw)); err != nil {
		return nil, err
	}
	return &orig, nil
}

func restoreMode(fd uintptr, orig *termios) {
	_ = ioctlPtr(fd, tcsets, unsafe.Pointer(orig))
}

// restoreOnSignal makes sure Ctrl+C (or a kill) during the menu doesn't
// leave the terminal stuck in raw mode. Call the returned cancel func once
// the menu returns normally.
func restoreOnSignal(fd uintptr, orig *termios) (cancel func()) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		select {
		case <-sigCh:
			restoreMode(fd, orig)
			os.Exit(130)
		case <-done:
		}
	}()
	return func() {
		signal.Stop(sigCh)
		close(done)
	}
}
