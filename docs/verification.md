# Verification

Last local verification in the sandbox:

```text
PASS shell syntax
PASS go test ./...
PASS go vet ./...
PASS go build
PASS hyprglass system --json
PASS Hyprland window/layer rule checks
PASS current hyprpaper wallpaper-block syntax
PASS Waybar CSS rejects 8-digit GTK-incompatible hex colors
PASS wallpaper and top bar config chain
PASS temp configs-only install
PASS no broken symlinks
PASS executable scripts
PASS duplicate package check for Arch and CachyOS profiles
PASS weak-leftover scan
PASS CachyOS dry-run package profile smoke test
WARN shellcheck missing in sandbox
WARN pacman missing in sandbox, so Arch/CachyOS package verification was skipped
SKIP runtime Hyprland compositor checks because the sandbox is not inside Hyprland
FINAL: WARN
```

The warning state is expected in this non-Arch, non-Hyprland sandbox. Re-run on the target laptop with:

```sh
./scripts/check.sh
```

On CachyOS, also run:

```sh
./install.sh --dry-run --distro=cachyos --rate-mirrors
hyprglass system --json
```
