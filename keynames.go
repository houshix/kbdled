package main

import "fmt"

// keyName maps a keycode to a readable name for display, falling back to
// the raw code. Covers common keys, not exhaustive.
func keyName(code uint16) string {
	if name, ok := keyNames[code]; ok {
		return name
	}
	return fmt.Sprintf("code %d", code)
}

var keyNames = map[uint16]string{
	1: "Esc",
	2: "1", 3: "2", 4: "3", 5: "4", 6: "5", 7: "6", 8: "7", 9: "8", 10: "9", 11: "0",
	12: "-", 13: "=", 14: "Backspace", 15: "Tab",
	16: "Q", 17: "W", 18: "E", 19: "R", 20: "T", 21: "Y", 22: "U", 23: "I", 24: "O", 25: "P",
	26: "[", 27: "]", 28: "Enter", 29: "Left Ctrl",
	30: "A", 31: "S", 32: "D", 33: "F", 34: "G", 35: "H", 36: "J", 37: "K", 38: "L",
	39: ";", 40: "'", 41: "`", 42: "Left Shift", 43: "\\",
	44: "Z", 45: "X", 46: "C", 47: "V", 48: "B", 49: "N", 50: "M", 51: ",", 52: ".", 53: "/",
	54: "Right Shift", 55: "KP *", 56: "Left Alt", 57: "Space", 58: "Caps Lock",
	59: "F1", 60: "F2", 61: "F3", 62: "F4", 63: "F5", 64: "F6",
	65: "F7", 66: "F8", 67: "F9", 68: "F10",
	69: "Num Lock", 70: "Scroll Lock",
	71: "KP 7", 72: "KP 8", 73: "KP 9", 74: "KP -",
	75: "KP 4", 76: "KP 5", 77: "KP 6", 78: "KP +",
	79: "KP 1", 80: "KP 2", 81: "KP 3", 82: "KP 0", 83: "KP .",
	87: "F11", 88: "F12",
	96: "KP Enter", 97: "Right Ctrl", 98: "KP /", 99: "Print Screen", 100: "Right Alt",
	102: "Home", 103: "Up", 104: "Page Up", 105: "Left", 106: "Right",
	107: "End", 108: "Down", 109: "Page Down", 110: "Insert", 111: "Delete",
	125: "Left Meta", 126: "Right Meta", 127: "Menu",
}
