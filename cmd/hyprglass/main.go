package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"hyprglass/internal/audio"
	"hyprglass/internal/bluetooth"
	"hyprglass/internal/command"
	"hyprglass/internal/display"
	"hyprglass/internal/doctor"
	"hyprglass/internal/lte"
	"hyprglass/internal/tui"
	"hyprglass/internal/wifi"
)

var (
	version    = "0.1.0"
	sourceRoot = ""
)

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
			root := findSourceRoot()
			if root == "" {
				fmt.Println("wallpaper generation requires a Hyprglass source checkout")
				os.Exit(1)
			}
			script := filepath.Join(root, "scripts", "generate-wallpaper.py")
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
	root := findSourceRoot()
	if root == "" {
		fmt.Println("Hyprglass could not find its source checkout, so it cannot auto-update. Run: sudo pacman -Syu, then rerun hyprglass doctor")
		return
	}
	script := filepath.Join(root, "install.sh")
	if _, err := os.Stat(script); err != nil {
		fmt.Println("Hyprglass could not find install.sh in its source checkout, so it cannot auto-update. Run: sudo pacman -Syu, then rerun hyprglass doctor")
		return
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		fmt.Println("bash is required to run the Hyprglass updater:", err)
		os.Exit(1)
	}
	fmt.Println("Running:", script, "--update")
	if err := syscall.Exec(bash, []string{bash, script, "--update"}, os.Environ()); err != nil {
		fmt.Println("could not start updater:", err)
		os.Exit(1)
	}
}
func repair() {
	msg := `Hyprglass repair runs only non-destructive checks in V0.
Targeted restarts you may choose manually:
  systemctl --user restart xdg-desktop-portal xdg-desktop-portal-hyprland
  pkill waybar; waybar &
  pkill mako; mako &
  sudo systemctl restart NetworkManager
  sudo systemctl restart bluetooth
  sudo systemctl restart ModemManager
Networking restarts are not automatic because they can drop your current session.`
	fmt.Println(strings.TrimSpace(msg))
}

func findSourceRoot() string {
	var candidates []string
	add := func(p string) {
		if p != "" {
			candidates = append(candidates, p)
		}
	}
	add(sourceRoot)
	if home, err := os.UserHomeDir(); err == nil {
		if b, err := os.ReadFile(filepath.Join(home, ".config", "hyprglass", "source-root")); err == nil {
			add(strings.TrimSpace(string(b)))
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		for {
			add(cwd)
			parent := filepath.Dir(cwd)
			if parent == cwd {
				break
			}
			cwd = parent
		}
	}
	if exe, err := os.Executable(); err == nil {
		add(filepath.Dir(exe))
	}
	for _, c := range candidates {
		if isSourceRoot(c) {
			return c
		}
	}
	return ""
}

func isSourceRoot(dir string) bool {
	if dir == "" {
		return false
	}
	for _, p := range []string{"install.sh", "cmd/hyprglass/main.go"} {
		if _, err := os.Stat(filepath.Join(dir, p)); err != nil {
			return false
		}
	}
	return true
}
