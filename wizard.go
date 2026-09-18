package main

// runWizard is the bare `kbdled` menu: Install is always available;
// Remap/Uninstall are disabled until something is installed.
func runWizard() {
	items := []menuItem{
		{label: "Install"},
		{label: "Remap", disabled: !isInstalled(), note: "run Install first"},
		{label: "Uninstall", disabled: !isInstalled(), note: "run Install first"},
		{label: "Exit"},
	}

	switch items[selectMenu(items)].label {
	case "Install":
		cmdInstall()
	case "Remap":
		cmdRemap()
	case "Uninstall":
		cmdUninstall()
	}
}

func isInstalled() bool {
	_, err := loadConfig()
	return err == nil
}
