package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"hyprglass/internal/audio"
	"hyprglass/internal/bluetooth"
	"hyprglass/internal/command"
	"hyprglass/internal/display"
	"hyprglass/internal/doctor"
	"hyprglass/internal/lte"
	"hyprglass/internal/tui"
	"hyprglass/internal/wifi"
)

const version = "0.1.0"

func main() {
	r := command.RealRunner{}
	args := os.Args[1:]
	if len(args) == 0 || args[0] == "--help" || args[0] == "help" {
		help()
		return
	}
	switch args[0] {
	case "version":
		fmt.Println("hyprglass", version)
	case "doctor":
		res := doctor.Run(r)
		if contains(args[1:], "--json") {
			b, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(b))
			return
		}
		tui.PrintChecks(res)
	case "wifi":
		wifi.RunTUI(r)
	case "bluetooth":
		if contains(args[1:], "--status") {
			bluetooth.PrintWaybarStatus(r)
			return
		}
		bluetooth.RunTUI(r)
	case "lte":
		lte.RunTUI(r)
	case "audio":
		audio.RunTUI(r)
	case "display":
		display.RunTUI(r)
	case "update":
		update()
	case "repair":
		repair()
	case "wallpaper":
		if len(args) > 1 && args[1] == "generate" {
			script := filepath.Join(repoRoot(), "scripts", "generate-wallpaper.py")
			out, err := r.Run("python3", script)
			if err != nil {
				fmt.Println("wallpaper generation failed:", err)
				fmt.Print(out)
				os.Exit(1)
			}
			fmt.Print(out)
			return
		}
		help()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		help()
		os.Exit(2)
	}
}
func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
func repoRoot() string { cwd, _ := os.Getwd(); return cwd }
func help() {
	fmt.Print(`Hyprglass is a fast, glassy Wayland desktop built on Hyprland for focused work.

Usage:
  hyprglass --help
  hyprglass version
  hyprglass doctor [--json]
  hyprglass wifi | bluetooth | lte | audio | display
  hyprglass update
  hyprglass repair
  hyprglass wallpaper generate
`)
}
func update() {
	fmt.Println("Hyprglass V0 does not auto-update without consent. Run: sudo pacman -Syu, then rerun hyprglass doctor")
}
func repair() {
	fmt.Println(strings.TrimSpace(`Hyprglass repair runs only non-destructive checks in V0.
Targeted restarts you may choose manually:
  systemctl --user restart xdg-desktop-portal xdg-desktop-portal-hyprland
  pkill waybar; waybar &
  pkill mako; mako &
  sudo systemctl restart NetworkManager
  sudo systemctl restart bluetooth
  sudo systemctl restart ModemManager
Networking restarts are not automatic because they can drop your current session.`))
}
