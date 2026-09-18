package main

import "fmt"

var lang = "en"

// pickLanguage shows the language picker with each name written in its own
// language, so people can find theirs regardless of the current default.
// If a config already exists (remap/uninstall on an already-installed
// system), it pre-selects the previously chosen language instead of
// always defaulting to English, and remembers whatever is picked.
func pickLanguage() {
	names := []string{"English", "Português", "Español", "Deutsch", "Français", "中文"}
	codes := []string{"en", "pt", "es", "de", "fr", "zh"}

	initial := 0
	if cfg, err := loadConfig(); err == nil {
		for i, c := range codes {
			if c == cfg.Lang {
				initial = i
				break
			}
		}
	}

	fmt.Println("Select language / Selecione o idioma / Seleccione el idioma /")
	fmt.Println("Sprache wählen / Choisir la langue / 选择语言")
	idx := selectMenu(names, initial)
	lang = codes[idx]

	if cfg, err := loadConfig(); err == nil {
		cfg.Lang = lang
		_ = saveConfig(cfg)
	}
}

// t looks up key in the current language, falling back to English if the
// language or the key is missing. args are applied with fmt.Sprintf only
// when present, so plain strings with no verbs are never re-parsed.
func t(key string, args ...interface{}) string {
	m, ok := translations[lang]
	if !ok {
		m = translations["en"]
	}
	s, ok := m[key]
	if !ok {
		s = translations["en"][key]
	}
	if len(args) > 0 {
		return fmt.Sprintf(s, args...)
	}
	return s
}

var translations = map[string]map[string]string{
	"en": {
		"press_key":       "Press the key you want to use to toggle the LED (%d seconds)...",
		"key_captured":    "Key captured: code %d on %s",
		"no_input_device": "No input device found under /dev/input",
		"no_key_detected": "No key press detected within %s",
		"no_stable_path":  "warning: no stable path (by-id/by-path) found for %s; the device name may change across reboots",
		"no_led_found":    "warning: no matching LED found under /sys/class/leds (looking for *::scrolllock); the key will be captured, but nothing will light up until a matching LED exists",
		"shortcut_note":   "If this key is currently bound to a shortcut in your desktop environment, remove that binding - this program now captures the key directly, and both would otherwise fire together.",
		"install_done":    "Installation complete. The key now works even on the login screen (SDDM/GDM/LightDM).",
		"uninstall_done":  "Removed: service stopped, LED reset, all files deleted.",
		"remap_prompt":    "Press the new key you want to use (%d seconds)...",
		"remap_done":      "Key updated. Restarting the service...",
		"config_missing":  "No configuration found. Run 'sudo kbdled install' first.",
	},
	"pt": {
		"press_key":       "Pressione a tecla que deseja usar para alternar o LED (%d segundos)...",
		"key_captured":    "Tecla capturada: código %d em %s",
		"no_input_device": "Nenhum dispositivo de entrada encontrado em /dev/input",
		"no_key_detected": "Nenhuma tecla detectada em %s",
		"no_stable_path":  "aviso: nenhum caminho estável (by-id/by-path) encontrado para %s; o nome do dispositivo pode mudar entre reboots",
		"no_led_found":    "aviso: nenhum LED correspondente encontrado em /sys/class/leds (procurando por *::scrolllock); a tecla será capturada, mas nada vai acender até existir um LED correspondente",
		"shortcut_note":   "Se essa tecla já estiver associada a um atalho no seu ambiente gráfico, remova essa associação - este programa agora captura a tecla diretamente, e os dois disparariam juntos.",
		"install_done":    "Instalação concluída. A tecla agora funciona até na tela de login (SDDM/GDM/LightDM).",
		"uninstall_done":  "Removido: serviço parado, LED resetado, todos os arquivos apagados.",
		"remap_prompt":    "Pressione a nova tecla que deseja usar (%d segundos)...",
		"remap_done":      "Tecla atualizada. Reiniciando o serviço...",
		"config_missing":  "Nenhuma configuração encontrada. Rode 'sudo kbdled install' primeiro.",
	},
	"es": {
		"press_key":       "Pulsa la tecla que quieres usar para alternar el LED (%d segundos)...",
		"key_captured":    "Tecla capturada: código %d en %s",
		"no_input_device": "No se encontró ningún dispositivo de entrada en /dev/input",
		"no_key_detected": "No se detectó ninguna tecla en %s",
		"no_stable_path":  "aviso: no se encontró una ruta estable (by-id/by-path) para %s; el nombre del dispositivo puede cambiar entre reinicios",
		"no_led_found":    "aviso: no se encontró ningún LED coincidente en /sys/class/leds (buscando *::scrolllock); la tecla será capturada, pero nada se encenderá hasta que exista un LED coincidente",
		"shortcut_note":   "Si esta tecla ya está asociada a un atajo en tu entorno de escritorio, elimina esa asociación - este programa ahora captura la tecla directamente, y ambos se activarían a la vez.",
		"install_done":    "Instalación completa. La tecla ahora funciona incluso en la pantalla de inicio de sesión (SDDM/GDM/LightDM).",
		"uninstall_done":  "Eliminado: servicio detenido, LED restablecido, todos los archivos borrados.",
		"remap_prompt":    "Pulsa la nueva tecla que quieres usar (%d segundos)...",
		"remap_done":      "Tecla actualizada. Reiniciando el servicio...",
		"config_missing":  "No se encontró ninguna configuración. Ejecuta 'sudo kbdled install' primero.",
	},
	"de": {
		"press_key":       "Drücke die Taste, mit der du die LED umschalten möchtest (%d Sekunden)...",
		"key_captured":    "Taste erfasst: Code %d auf %s",
		"no_input_device": "Kein Eingabegerät unter /dev/input gefunden",
		"no_key_detected": "Innerhalb von %s wurde kein Tastendruck erkannt",
		"no_stable_path":  "Warnung: kein stabiler Pfad (by-id/by-path) für %s gefunden; der Gerätename kann sich nach einem Neustart ändern",
		"no_led_found":    "Warnung: keine passende LED unter /sys/class/leds gefunden (gesucht wird *::scrolllock); die Taste wird erfasst, aber nichts leuchtet auf, bis eine passende LED existiert",
		"shortcut_note":   "Falls diese Taste bereits mit einer Tastenkombination in deiner Desktop-Umgebung verknüpft ist, entferne diese Verknüpfung - dieses Programm erfasst die Taste jetzt direkt, sonst würden beide gleichzeitig auslösen.",
		"install_done":    "Installation abgeschlossen. Die Taste funktioniert jetzt auch am Anmeldebildschirm (SDDM/GDM/LightDM).",
		"uninstall_done":  "Entfernt: Dienst gestoppt, LED zurückgesetzt, alle Dateien gelöscht.",
		"remap_prompt":    "Drücke die neue Taste, die du verwenden möchtest (%d Sekunden)...",
		"remap_done":      "Taste aktualisiert. Dienst wird neu gestartet...",
		"config_missing":  "Keine Konfiguration gefunden. Führe zuerst 'sudo kbdled install' aus.",
	},
	"fr": {
		"press_key":       "Appuyez sur la touche que vous voulez utiliser pour basculer la LED (%d secondes)...",
		"key_captured":    "Touche capturée : code %d sur %s",
		"no_input_device": "Aucun périphérique d'entrée trouvé dans /dev/input",
		"no_key_detected": "Aucune touche détectée en %s",
		"no_stable_path":  "avertissement : aucun chemin stable (by-id/by-path) trouvé pour %s ; le nom de l'appareil peut changer après un redémarrage",
		"no_led_found":    "avertissement : aucune LED correspondante trouvée dans /sys/class/leds (recherche de *::scrolllock) ; la touche sera capturée, mais rien ne s'allumera tant qu'aucune LED correspondante n'existe",
		"shortcut_note":   "Si cette touche est déjà associée à un raccourci dans votre environnement de bureau, supprimez cette association - ce programme capture désormais la touche directement, et les deux se déclencheraient ensemble.",
		"install_done":    "Installation terminée. La touche fonctionne désormais même sur l'écran de connexion (SDDM/GDM/LightDM).",
		"uninstall_done":  "Supprimé : service arrêté, LED réinitialisée, tous les fichiers effacés.",
		"remap_prompt":    "Appuyez sur la nouvelle touche que vous voulez utiliser (%d secondes)...",
		"remap_done":      "Touche mise à jour. Redémarrage du service...",
		"config_missing":  "Aucune configuration trouvée. Exécutez d'abord 'sudo kbdled install'.",
	},
	"zh": {
		"press_key":       "请按下你想用来切换 LED 的按键(%d 秒内有效)...",
		"key_captured":    "已捕获按键:代码 %d,设备 %s",
		"no_input_device": "在 /dev/input 下未找到任何输入设备",
		"no_key_detected": "在 %s 内未检测到按键",
		"no_stable_path":  "警告:未找到 %s 的稳定路径(by-id/by-path);设备名称可能在重启后发生变化",
		"no_led_found":    "警告:在 /sys/class/leds 下未找到匹配的 LED(查找 *::scrolllock);按键会被捕获,但在出现匹配的 LED 之前不会有任何指示灯亮起",
		"shortcut_note":   "如果该按键已绑定到桌面环境中的快捷键,请先移除该绑定——本程序现在会直接捕获按键,否则两者会同时触发。",
		"install_done":    "安装完成。该按键现在即使在登录界面(SDDM/GDM/LightDM)也能使用。",
		"uninstall_done":  "已移除:服务已停止,LED 已重置,所有文件已删除。",
		"remap_prompt":    "请按下你想使用的新按键(%d 秒内有效)...",
		"remap_done":      "按键已更新,正在重启服务...",
		"config_missing":  "未找到配置。请先运行 'sudo kbdled install'。",
	},
}
