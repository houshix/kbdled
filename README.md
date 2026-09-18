# kbdled

✨ kbdled lets you toggle a keyboard LED (Scroll Lock, by default) with a dedicated key — even on the login screen, before you've signed in. ✨

⚠️ » kbdled reads your keyboard's raw input device and needs to run as root. See [Important Notes](#️-important-notes) before installing.

🤔 What do you mean by "even on the login screen"?

A shortcut bound inside GNOME/KDE/etc. only exists while that session is open — it can never fire at SDDM/GDM/LightDM, since those run as a separate process with no access to your session at all. kbdled captures the key one layer below any session, straight from the kernel's input device, so it works everywhere: login screen, TTY, any desktop.

🤔 What else does it do?

 Interactive arrow-key wizard (Space, Enter, or numpad Enter to confirm) — just run `kbdled` with no arguments;
 Binds a single key or a key combination (hold two or more keys together, then release);
 Change the bound key later (`remap`) without a full reinstall;
 Survives the keyboard being unplugged and reconnected, or missing at boot;
 Fixes the kernel resyncing Caps/Num Lock and stomping on your LED;
 Single static binary — no runtime, no third-party Go modules, no extra daemons (`udev` rules, `actkbd`, etc. not needed).

🤔 How to install?

**Option A — prebuilt binary (fastest)**

1. Grab the latest `kbdled.zip` from [Releases](https://github.com/houshix/kbdled/releases).
2. Unzip it and run:
```bash
unzip kbdled.zip
sudo ./kbdled
```

**Option B — build from source**
```bash
sudo apt install golang-go   # or your distro's equivalent
git clone https://github.com/houshix/kbdled.git
cd kbdled
go build -o kbdled .
sudo ./kbdled
```

Running `kbdled` with no arguments drops you into the wizard: pick **Install**, press
and hold the key (or key combination) you want to use, then release it. Remap and
Uninstall show up grayed out until something is actually installed.

🎛️ Usage

Prefer typing over navigating a menu? Every wizard option is also a direct command:
```bash
sudo kbdled install     # first-time setup (see above)
sudo kbdled remap       # change the key/combination without a full reinstall
sudo kbdled uninstall   # stop the service, reset the LED, remove all files
```

⚙️ How it works

 The daemon reads the configured `/dev/input` device directly and is started by `multi-user.target`, which comes up before display managers (`graphical.target`) do — so it's already active by the time the login screen appears;
 Caps/Num Lock changing state makes the kernel resync all three lock LEDs together, at a level the LED "trigger" mechanism can't touch. kbdled listens for that directly (`EV_LED` events on the same device) and puts its own LED state back immediately, backed by a 2s safety-net poll;
 If the keyboard isn't there yet (still booting, unplugged), it waits and retries instead of exiting, so it never needs `systemd`'s restart limit to recover.

🗒️ Important Notes

 **Needs root** — it reads a raw input device and writes to systemd/sysfs;
 **Doesn't grab the key exclusively** (no `EVIOCGRAB`) — it still reaches your desktop session too, so remove any existing shortcut bound to the same key;
 **Binds to one keyboard** — whichever one you press the key on during `install`/`remap`;
 **`Fn` probably won't work** — it's usually consumed by the keyboard's own firmware before the OS ever sees a key event for it, so a combination involving `Fn` likely won't register;
 **64-bit only** — the raw event parsing assumes the 64-bit `input_event` layout;
 Defaults to the Scroll Lock LED; edit `scrollLockLEDs` in `led.go` and rebuild to target Caps or Num Lock instead.

Extras:

Found a bug or have an idea? Open an issue or send a pull request — contributions welcome.
