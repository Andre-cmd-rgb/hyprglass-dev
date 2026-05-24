# Verification

Last local verification in the sandbox for V1.4:

```text
PASS shell syntax
PASS go test ./...
PASS go vet ./...
PASS go build
PASS hyprglass --help
PASS hyprglass system --json
PASS hyprglass settings close path
PASS hyprglass laptop/power menu smoke tests
PASS doctor JSON generation
PASS Hyprland source checks
PASS Hyprland window/layer rule checks
PASS current hyprpaper wallpaper-block syntax
PASS Hyprglass managed display block exists
PASS Settings display page uses display-scoped apply logic
PASS installer has display-preservation guard
PASS Waybar detach/reload checks
PASS Waybar CSS avoids 8-digit GTK-incompatible hex colors
PASS Waybar icon font fallback chain
PASS temp configs-only install
PASS temp install wallpaper/top-bar chain
PASS installer preserves existing display config
PASS no broken symlinks
PASS executable scripts
PASS duplicate package check for Arch and CachyOS profiles
PASS weak-leftover scan
WARN shellcheck missing in sandbox
WARN pacman missing in sandbox, so Arch/CachyOS package verification was skipped
SKIP runtime Hyprland compositor checks because the sandbox is not inside Hyprland
FINAL: WARN
```

The warning state is expected in this non-Arch, non-Hyprland sandbox. Re-run on the target laptop with:

```sh
./scripts/check.sh
```

Before installing on your actual PC, run:

```sh
./install.sh --dry-run
```

On CachyOS, also run:

```sh
./install.sh --dry-run --distro=cachyos --rate-mirrors
hyprglass system --json
```
