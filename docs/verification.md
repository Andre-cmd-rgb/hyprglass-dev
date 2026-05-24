# Verification

Last local verification in the sandbox:

```text
PASS shell syntax
PASS go test ./...
PASS go vet ./...
PASS go build
PASS Hyprland window/layer rule checks
PASS current hyprpaper wallpaper-block syntax
PASS Waybar CSS rejects 8-digit GTK-incompatible hex colors
PASS wallpaper and top bar config chain
PASS temp configs-only install
PASS no broken symlinks
PASS executable scripts
PASS duplicate package check
PASS weak-leftover scan
WARN shellcheck missing in sandbox
WARN pacman missing in sandbox, so Arch package verification was skipped
SKIP runtime Hyprland compositor checks because the sandbox is not inside Hyprland
FINAL: WARN
```

The warning state is expected in this non-Arch, non-Hyprland sandbox. Re-run on the target Arch laptop with:

```sh
./scripts/check.sh
```
