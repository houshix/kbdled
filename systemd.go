package main

import "os"

const (
	binaryPath  = "/usr/local/bin/kbdled"
	unitPath    = "/etc/systemd/system/kbdled.service"
	serviceName = "kbdled.service"
)

// No After=: combined with WantedBy=multi-user.target it would form an
// ordering cycle, which systemd breaks by dropping After= anyway.
// WantedBy alone starts this alongside multi-user.target, before display
// managers (graphical.target).
const unitContent = `[Unit]
Description=Keyboard LED Toggle Daemon
StartLimitIntervalSec=30
StartLimitBurst=5

[Service]
Type=simple
ExecStart=` + binaryPath + ` daemon
Restart=always
RestartSec=1
User=root

[Install]
WantedBy=multi-user.target
`

func writeUnit() error {
	return os.WriteFile(unitPath, []byte(unitContent), 0644)
}
