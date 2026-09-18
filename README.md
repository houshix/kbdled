# kbdled

Toggle a keyboard LED (scroll lock, by default) with a dedicated key —
including on the login screen (SDDM/GDM/LightDM), because it reads the
input device directly instead of relying on a desktop shortcut.

Single static binary, Go standard library only (zero external
dependencies), with an interactive setup wizard in six languages.

## Features

- Works everywhere: login screen, TTY, any desktop session — the key is
  captured from `/dev/input`, not from a desktop environment's shortcut
  system.
- One binary does everything: `install`, `remap`, `uninstall`, and the
  background daemon are all the same executable.
- Interactive wizard with arrow-key menus (Up/Down + Enter, Esc to cancel),
  available in English, Português, Español, Deutsch, Français and 中文.
  Remembers your chosen language for next time.
- No third-party Go modules, no `udev` rules, no extra system daemons
  (`actkbd` and similar are not needed).
- Change the bound key later with `remap`, without touching the rest of
  the setup.

## Installation

Requires only the Go compiler (no modules to download — `go.mod` lists no
dependencies):

```bash
sudo apt install golang-go   # or your distro's equivalent
git clone https://github.com/<you>/kbdled.git
cd kbdled
go build -o kbdled .
sudo ./kbdled install
```

`install` will:

1. Ask you to pick a language.
2. Ask you to press the key you want to use (20 second window).
3. Save that key's device path and keycode to `/etc/kbdled/config.json`.
4. Copy itself to `/usr/local/bin/kbdled`.
5. Create and enable a systemd service that starts before the login
   screen.

If no matching LED or no `systemctl` is found, it says so instead of
silently reporting success.

## Usage

```bash
sudo kbdled install     # first-time setup (see above)
sudo kbdled remap       # change the key without a full reinstall
sudo kbdled uninstall   # stop the service, reset the LED, remove all files
```

`kbdled daemon` also exists, but it's meant to be invoked by the systemd
unit, not run by hand.

## How it works

**The problem this solves:** a keyboard shortcut configured inside a
desktop environment (KDE, GNOME, etc.) only exists while that session is
running and focused. The SDDM/GDM/LightDM login screen is a separate
process with no access to any session's shortcut configuration, so a
DE-level shortcut can never fire there. The fix is to capture the key one
layer below any session, directly from the kernel's input device.

**Key capture.** `install` and `remap` open every `/dev/input/eventN` at
once and wait for the first plausible keyboard key press (event type
`EV_KEY`, value `1`, code below the `BTN_*` range used by mice/joysticks —
this also excludes a handful of multimedia/macro keys that live above that
range). Whichever device reports it first is resolved to a stable path
under `/dev/input/by-id/` (falling back to `/dev/input/by-path/`) so the
configuration survives reboots even if the `eventN` number changes.

**The daemon.** A single long-running process, installed as
`kbdled.service`, pulled in by `multi-user.target` — which display
managers (SDDM/GDM/LightDM) start after, as part of `graphical.target` —
so the daemon is already active by the time the login screen appears.

Caps Lock and Num Lock changing state makes the kernel resync all three
lock LEDs together (this happens in a lower-level input-core path, not
through the swappable "trigger" mechanism, so there's no sysfs switch that
turns it off), which clobbers whatever this program last set. The daemon
deals with this by listening for `EV_LED` events on the same device it's
already reading for the toggle key — any such event means "the kernel just
touched an LED here," so it checks its own LED against the desired state
and puts it back if the kernel changed it. A low-frequency timer (every 2s)
backs this up in case a particular driver doesn't surface a clean event for
some resync path. It also disables the LED's own trigger (`echo none >
.../trigger`) on top of this, mostly as a no-cost precaution against
anything else that might independently be driving it.

The chosen state is persisted to `/run/kbdled.state` (tmpfs, not
world-writable, cleared on reboot) so a restart doesn't lose it.
`uninstall` restores whatever trigger the kernel originally had on the LED.

If the configured keyboard isn't present - unplugged, or not yet enumerated
this early in boot - the daemon waits and keeps retrying every 2 seconds
rather than exiting; the same recovery path handles it being unplugged and
reconnected later. This matters because exiting and leaning on systemd's
`Restart=always` to bring it back can run into the restart-rate limit
within seconds if the keyboard stays missing for a while, leaving the
service stuck in a failed state until the next reboot.

**The interactive menu.** Arrow-key selection is implemented with a
direct `ioctl` on the standard `TCGETS`/`TCSETS` calls (raw terminal mode)
rather than pulling in `golang.org/x/term`, to keep the dependency count
at zero. This targets mainstream Linux (x86/ARM), which covers the
realistic audience for a Linux keyboard-LED tool. On a non-interactive
stdin (e.g. piped input), the wizard automatically falls back to a plain
numbered prompt.

## Files this program manages

| Path | Purpose |
|---|---|
| `/usr/local/bin/kbdled` | the binary |
| `/etc/kbdled/config.json` | device, keycode, chosen language, original LED trigger |
| `/etc/systemd/system/kbdled.service` | the systemd unit |
| `/run/kbdled.state` | current LED state (cleared on reboot) |

`uninstall` restores the LED's original trigger and removes all four.

## Customization

By default this controls the scroll-lock LED
(`/sys/class/leds/*::scrolllock`). To target Caps Lock or Num Lock
instead, change the glob pattern in `led.go` (`scrollLockLEDs`) and
rebuild.

## Known limitations

- **One input device.** Whichever keyboard you press the key on during
  `install`/`remap` is the one the daemon listens to afterward. If you
  regularly switch between a laptop's built-in keyboard and an external
  one, only the one you used during setup will trigger the toggle.
- **The key isn't "eaten."** This doesn't grab the device exclusively
  (`EVIOCGRAB`), so the key press still reaches your desktop session
  normally. If it's also bound to a shortcut there, remove that binding —
  otherwise both would fire.
- **The daemon can see the whole keystream of that device.** It only acts
  on the one configured keycode, but since it reads the raw device file,
  every key press on that keyboard passes through the process (running as
  root) to get there. It doesn't log or store anything beyond the current
  LED state, but this is worth knowing given what it's granted access to.
- **64-bit only.** The raw `input_event` parsing assumes the 24-byte
  layout used on 64-bit Linux; it will misparse events on a 32-bit kernel.
