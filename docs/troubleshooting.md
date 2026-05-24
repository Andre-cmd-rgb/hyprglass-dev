# Troubleshooting

Run `hyprglass doctor` first. For machine-readable output, run `hyprglass doctor --json`.

Hyprland does not start: check that `hyprland`, `xdg-desktop-portal`, and `xdg-desktop-portal-hyprland` are installed. Start from a TTY with `Hyprland` and read the printed config error.

Waybar missing: run `waybar` in a terminal and check `~/.config/waybar/config.jsonc` and `style.css`.

Black screen: from TTY, restore `~/.config/hypr/conf.d/monitors.conf` from the backup folder or set `monitor = , preferred, auto, 1`.

Portals or screen sharing broken: restart user portals with `systemctl --user restart xdg-desktop-portal xdg-desktop-portal-hyprland` after confirming those units exist on your system.

Wi-Fi not listed: check `systemctl status NetworkManager`, `nmcli radio wifi`, then rescan with `nmcli device wifi rescan`.

Bluetooth not working: check `systemctl status bluetooth` and `bluetoothctl show`. Pairing can require an interactive agent.

LTE modem not detected: check `systemctl status ModemManager`, `mmcli -L`, and that the modem is not blocked by rfkill or firmware state.

APN issues: use the carrier APN exactly, then test with `mmcli -m <index> --simple-connect="apn=<apn>"`.

Audio missing: check `wpctl status`, `systemctl --user status pipewire wireplumber`, and confirm `pipewire-pulse` is installed.

Kitty too transparent: edit `~/.config/kitty/kitty.conf` and increase `background_opacity` toward `0.90`.

Restore backup configs: copy files back from `~/.config/hyprglass-backups/<timestamp>/`.
