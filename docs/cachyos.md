# CachyOS support

Hyprglass supports both clean Arch Linux and CachyOS.

## What support means

- The installer auto-detects CachyOS from `/etc/os-release`.
- `./install.sh --distro=cachyos` forces the CachyOS package profile.
- CachyOS installs use `packages/cachyos-core.txt`.
- `hyprglass system` exposes CachyOS-aware maintenance actions.
- `hyprglass doctor` checks for CachyOS tools when the host is CachyOS.

## What Hyprglass does not do

Hyprglass does not convert Arch into CachyOS, does not rewrite pacman repositories, and does not install or replace kernels without explicit user action. That is deliberate. CachyOS repository migration and kernel selection are OS-level changes, not rice-level changes.

## Useful commands

```sh
./install.sh --distro=cachyos
./install.sh --distro=cachyos --rate-mirrors
hyprglass system
hyprglass system --json
hyprglass system rate-mirrors
hyprglass system chwd-list
hyprglass system chwd-auto
```

`rate-mirrors`, `chwd-auto`, and `pacman -Syu` actions are confirmation-gated.

## Performance stance

On CachyOS, Hyprglass should stay thin and let CachyOS provide the speed: optimized repositories, CachyOS kernel choices, mirror tooling, and hardware detection. Hyprglass should provide a clean Hyprland laptop workstation layer on top, not fight the distribution.

## V1.2 hardware setup behavior

Hyprglass can run CachyOS hardware auto-configuration during install only when explicitly requested:

```sh
./install.sh --distro=cachyos --auto-hardware
```

This calls `sudo chwd -a` after package install if CachyOS is detected and `chwd` exists. It is opt-in because driver/profile changes are OS-level changes, not cosmetic rice changes.
