package main

import "os"

const (
	binaryPath  = "/usr/local/bin/kbdled"
	unitPath    = "/etc/systemd/system/kbdled.service"
	serviceName = "kbdled.service"
)

// No After= here on purpose: After=multi-user.target combined with
// WantedBy=multi-user.target is an ordering cycle (systemd breaks it by
// dropping the After=, which silently defeats the "be up before login"
// goal). WantedBy alone already starts this at the same point in boot as
// everything else pulled in by multi-user.target - which is what we want,
// since display managers (SDDM/GDM/LightDM) start later, in
// graphical.target.
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
