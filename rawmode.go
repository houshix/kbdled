package main

import (
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

// Raw-mode terminal handling for Linux (x86/ARM), reading input one byte
// at a time. Avoids golang.org/x/term (and its x/sys dependency) to keep
// the module dependency-free; this termios layout (for TCGETS/TCSETS) is
// stable across mainstream Linux architectures.
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

// enableRawMode disables canonical mode and echo for byte-at-a-time
// reads. Returns the original settings for restoreMode.
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

// restoreOnSignal restores the terminal on Ctrl+C/kill during the menu.
// Call the returned cancel func on normal return.
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
