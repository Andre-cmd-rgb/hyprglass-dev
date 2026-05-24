# Troubleshooting

Run `hyprglass doctor` first. For machine-readable output, run `hyprglass doctor --json`.

## Hyprland config errors on startup

If you see a red error bar listing config errors, the most common causes are:

- **`config option <X> does not exist`** — a deprecated option from an older Hyprland version.
  Run `hyprctl version` to confirm your version, then check the release notes.
- **`windowrulev2`** — removed in 0.53; replaced by `windowrule = <rule>, match:class <regex>`.
- **`gestures:workspace_swipe`** — removed in 0.51; replaced by `gesture = 3, horizontal, workspace`.
- **`dwindle:pseudotile`** — removed in 0.55; use `windowrule = pseudo, match:class <class>` per-app.

## Hyprland does not start

Check that `hyprland`, `xdg-desktop-portal`, and `xdg-desktop-portal-hyprland` are installed.
Start from a TTY with `start-hyprland` (0.53+) or `Hyprland` and read the printed config error.

## Wallpaper not showing

Hyprpaper needs `~/.config/hypr/hyprpaper.conf`. Confirm `exec-once = hyprpaper` is in
`autostart.conf` and that the wallpaper path in `hyprpaper.conf` exists.

## Waybar not showing

Run `waybar` in a terminal and check stderr. Common causes: missing font
(JetBrainsMono Nerd Font), JSON syntax error in `config.jsonc`, or a missing binary
referenced in an exec field.

## Lock screen blank / no background

`hyprlock` reads `~/.config/hypr/hyprlock.conf`, which points at
`~/.config/hypr/assets/wallpapers/hyprglass-dusk.png`. Confirm `install.sh`
was run and the wallpaper was copied. Run `ls ~/.config/hypr/assets/` to verify.

## `hyprglass: command not found`

The binary is installed to `~/.local/bin/hyprglass`. The installer writes PATH
setup to your shell startup file and `~/.profile`, but the current terminal cannot
inherit that change. Open a new terminal or run:

```
exec $SHELL -l
```

For the current terminal only:

```
export PATH="$HOME/.local/bin:$PATH"
```

## Black screen after login

From a TTY, restore `~/.config/hypr/conf.d/monitors.conf` from the backup folder, or reset it:
```
echo 'monitor = , preferred, auto, 1' > ~/.config/hypr/conf.d/monitors.conf
```

## Portals or screen sharing broken

```
systemctl --user restart xdg-desktop-portal xdg-desktop-portal-hyprland
```
Confirm those units exist on your system first.

## Volume keys not working

Check that `wireplumber` is running: `systemctl --user status wireplumber`.
Verify `wpctl status` shows a default sink. The Super+F9/F10/F11 fallback binds
work even in VMs where XF86 keys are not forwarded.

## Brightness keys not working

`brightnessctl` must be installed and your user must be in the `video` group:
```
sudo usermod -aG video $USER   # then log out and back in
brightnessctl set +5%           # test manually
```
On a VM, `brightnessctl` will print "No backlight found" — this is expected and harmless.

## Wi-Fi not listed

Check `systemctl status NetworkManager`, `nmcli radio wifi`, then rescan:
```
nmcli device wifi rescan
```

## Bluetooth not working

Check `systemctl status bluetooth` and `bluetoothctl show`. Pairing may require an
interactive agent; run `bluetoothctl` directly for that.

## LTE modem not detected

Check `systemctl status ModemManager`, `mmcli -L`, and that the modem is not blocked
by rfkill or firmware state. APN test: `mmcli -m <index> --simple-connect="apn=<apn>"`.

## Audio missing

Check `wpctl status`, `systemctl --user status pipewire wireplumber`, and confirm
`pipewire-pulse` is installed. If PulseAudio is also installed, it may conflict.

## Polkit prompts not appearing (sudo GUI / mount dialogs)

Confirm `polkit-gnome` is installed and the agent is running:
```
pgrep -a polkit
```
If missing, `exec-once = /usr/lib/polkit-gnome/polkit-gnome-authentication-agent-1`
must be in `autostart.conf`.

## Kitty too transparent

Edit `~/.config/kitty/kitty.conf` and increase `background_opacity` toward `0.92`.

## Restore backed-up configs

```
cp -a ~/.config/hyprglass-backups/<timestamp>/hypr ~/.config/hypr
# repeat for other dirs as needed
```
