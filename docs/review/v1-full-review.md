# Hyprglass V1.1 full repo review

## Verdict

V1 fixed the wallpaper and Waybar breakages, but the repository still had three product-level problems:

1. It only treated Arch as the real host even though CachyOS is a realistic target for the product.
2. Package installation was hardwired to one package list.
3. System maintenance was scattered; Settings did not expose distro/hardware/update information in one place.

V1.1 fixes those without bloating the project or turning Hyprglass into a distro.

## Changes made in this pass

- Added `internal/platform` for `/etc/os-release` parsing and Arch/CachyOS detection.
- Added `packages/cachyos-core.txt`.
- Added `hyprglass system` and `hyprglass cachyos` alias.
- Added System/CachyOS entry inside the main Settings app.
- Added CachyOS mirror ranking action using `cachyos-rate-mirrors` when available.
- Added CachyOS hardware profile list and confirmation-gated `chwd -a` action.
- Updated `install.sh` with `--distro=auto|arch|cachyos` and `--rate-mirrors`.
- Updated helper scripts so package installation uses the same distro profile logic.
- Updated Doctor to recognize CachyOS as supported instead of warning as generic non-Arch.
- Added tests for distro parsing and system status collection.
- Expanded verification to cover the new system command and both package profiles.

## Review notes

### Installer

The installer is now acceptable for Arch and CachyOS. It still intentionally does not enable CachyOS repositories on vanilla Arch. That would be an OS migration and should never happen as a side effect of installing a desktop configuration.

### Settings app

The Settings app is now the right place for appearance, display, keyboard, network, modem, power, services, updates, and distro/system maintenance. This is closer to the single Apple-like settings surface the project needs.

### Package lists

Arch and CachyOS profiles are separate. Most packages remain identical because CachyOS is Arch-compatible and ships the same package names for the Hyprland stack. CachyOS-specific tools are isolated to the CachyOS profile.

### Runtime gaps

These still require real hardware/session testing:

- Actual CachyOS package verification through pacman.
- Actual `cachyos-rate-mirrors` run.
- Actual `chwd` hardware profile behavior.
- Hyprland runtime reload behavior on a live session.
- Modem unlock on real LTE/5G hardware.

## Keep scope disciplined

Do not add gaming meta packages, profile switching, giant GUI settings, or kernel replacement automation. CachyOS already handles speed. Hyprglass should make the desktop clean, fast, coherent, and laptop-friendly.
