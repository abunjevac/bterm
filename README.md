# bterm

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](go.mod)
[![Platform](https://img.shields.io/badge/platform-Linux-lightgrey?logo=linux&logoColor=white)](https://github.com/abunjevac/bterm)
[![Release](https://img.shields.io/github/v/release/abunjevac/bterm)](https://github.com/abunjevac/bterm/releases)
[![Build](https://github.com/abunjevac/bterm/actions/workflows/build.yml/badge.svg)](https://github.com/abunjevac/bterm/actions/workflows/build.yml)

An opinionated GTK4 terminal emulator with tabs and pane splitting. Built with Go, GTK4, and VTE.

---

## Features

- **Tabs** — multiple terminals with numbered labels in the title bar
- **Pane splitting** — split any pane horizontally or vertically, infinitely nestable
- **Keyboard-first navigation** — focus and resize panes without the mouse
- **Font zoom** — adjust font size in the focused pane at runtime
- **CWD inheritance** — new tabs and splits open in the working directory of the active pane
- **Clipboard toasts** — brief overlay confirms copy and paste actions
- **Theme system** — TOML palette files; ships with a Dracula-compatible default
- **Configurable keybindings** — remap any action in `~/.config/bterm/keymap.toml`

---

## Requirements

- Linux
- GTK4 (`libgtk-4-dev`)
- VTE GTK4 (`libvte-2.91-gtk4-dev`)
- Go 1.26+
- [Task](https://taskfile.dev) (optional, for `task` commands)

Install system dependencies on Ubuntu/Debian:

```bash
sudo apt install libgtk-4-dev libvte-2.91-gtk4-dev
```

---

## Installation

### Download a release

Grab the latest Linux binary from the [Releases](../../releases) page and put it on your `$PATH`:

```bash
chmod +x bterm
sudo mv bterm /usr/local/bin/
```

### Build from source

```bash
git clone https://github.com/abunjevac/bterm.git
cd bterm
task build          # requires Task — https://taskfile.dev
# or: go build -o bterm ./cmd/bterm
```

### Desktop launcher and icon

Linux desktop shells resolve application icons through desktop integration metadata, not directly from a running GTK4
binary. The application window advertises the icon name `io.github.abunjevac.bterm`, while the desktop shell looks up
that name in the icon theme and associates the window with a `.desktop` entry. Without those files, the shell may show a
generic icon even though bterm has its own embedded artwork used in the About dialog.

Install the user-local desktop launcher and icon after installing or building the binary:

```bash
task install-desktop
```

This installs:

| File | Destination |
|------|-------------|
| Desktop entry | `~/.local/share/applications/io.github.abunjevac.bterm.desktop` |
| 512px icon | `~/.local/share/icons/hicolor/512x512/apps/io.github.abunjevac.bterm.png` |

The task also refreshes the desktop application database and icon cache. Close and relaunch bterm after installation so
the desktop shell can associate the new window with the launcher entry.

To remove the desktop integration files and refresh the caches:

```bash
task uninstall-desktop
```

---

## Configuration

bterm writes its configuration to `~/.config/bterm/` on first launch.

| File          | Purpose                             |
|---------------|-------------------------------------|
| `config.toml` | Font, theme, window size, shell     |
| `keymap.toml` | Key bindings                        |
| `themes/`     | Custom TOML palette files           |

Open `config.toml` from within bterm via **Open Config** in the hamburger menu or `Ctrl+,`.

### config.toml

```toml
font           = "Monospace"
font_size      = 12.0
theme          = "dracula"
shell          = ""          # defaults to $SHELL
shell_args     = ["-l"]
scrollback     = 10000
window_columns = 180
window_rows    = 40
title          = "bterm"
terminal_notification_method = "dbus"  # "dbus" or "off"
```

Terminal notifications are enabled by default. bterm listens for sequences such as `OSC 777;notify;Title;Message ST` and `OSC 9;Message ST`, then sends them directly to `org.freedesktop.Notifications` over D-Bus.

---

## Keyboard shortcuts

All bindings are configurable in `~/.config/bterm/keymap.toml`. The current set is also visible in the app via
**Keyboard Shortcuts** in the hamburger menu.

### Tabs

| Shortcut           | Action                |
|--------------------|-----------------------|
| `Ctrl+Shift+T`     | New tab at end        |
| `Ctrl+Shift+R`     | New tab after current |
| `Ctrl+Shift+Q`     | Close tab             |
| `Alt+1` – `Alt+9`  | Switch to tab N       |

### Panes

| Shortcut                  | Action                |
|---------------------------|-----------------------|
| `Ctrl+Shift+O`            | Split pane left/right |
| `Ctrl+Shift+E`            | Split pane top/bottom |
| `Ctrl+Shift+W`            | Close focused pane    |
| `Alt+←/→/↑/↓`            | Move focus to pane    |
| `Ctrl+Shift+←/→/↑/↓`     | Resize focused pane   |

### Clipboard

| Shortcut                         | Action |
|----------------------------------|--------|
| `Ctrl+Shift+C` / `Ctrl+Insert`   | Copy   |
| `Ctrl+Shift+V` / `Shift+Insert`  | Paste  |

### Font

| Shortcut       | Action     |
|----------------|------------|
| `Ctrl+Numpad+` | Zoom in    |
| `Ctrl+Numpad−` | Zoom out   |
| `Ctrl+Numpad*` | Reset zoom |

### App

| Shortcut | Action      |
|----------|-------------|
| `Ctrl+,` | Open Config |

---

## Releasing

Tag a commit to trigger an automated GitHub Actions build:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The workflow builds a Linux binary, creates a GitHub release with auto-generated notes, and prunes releases beyond the
three most recent.

---

## Development

```bash
task build    # build binary
task run      # build and run
task test     # run tests
task lint     # golangci-lint + go vet
task check    # lint + test + osv-scanner
task tidy     # go mod tidy
```

---

## License

[MIT](LICENSE) © Alan Bunjevac
