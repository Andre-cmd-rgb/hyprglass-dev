package doctor

import (
	"hyprglass/internal/command"
	"hyprglass/internal/platform"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Check struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Fix     string `json:"fix,omitempty"`
}
type Result struct {
	Status string  `json:"status"`
	Checks []Check `json:"checks"`
}
type requiredFile struct {
	RepoPath      string
	InstalledPath string
}

func Run(r command.Runner) Result {
	var cs []Check
	add := func(n, s, m, f string) { cs = append(cs, Check{Name: n, Status: s, Message: m, Fix: f}) }
	if runtime.GOOS != "linux" {
		add("operating system", "warn", "not running on Linux", "Run Hyprglass on Arch Linux")
	} else {
		add("operating system", "pass", "Linux detected", "")
	}
	info := platform.Read()
	if info.ArchLike {
		label := "Arch-like Linux detected"
		if info.CachyOS {
			label = "CachyOS detected"
		}
		add("distro", "pass", label+" ("+info.DisplayName()+")", "")
	} else {
		add("distro", "warn", "this environment does not look Arch-compatible", "Run on Arch Linux or CachyOS for package verification")
	}
	if info.CachyOS {
		if r.Exists("cachyos-rate-mirrors") {
			add("cachyos-rate-mirrors", "pass", "available", "")
		} else {
			add("cachyos-rate-mirrors", "warn", "missing", "Install/repair CachyOS tools or use CachyOS Hello")
		}
		if r.Exists("chwd") {
			add("chwd", "pass", "CachyOS hardware detection available", "")
		} else {
			add("chwd", "warn", "CachyOS hardware detection missing", "Install/repair chwd if hardware driver management is needed")
		}
	}
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		add("hyprland session", "pass", "Hyprland session variable present", "")
	} else {
		add("hyprland session", "warn", "not inside Hyprland; runtime compositor checks skipped", "Start Hyprland then rerun doctor")
	}
	for _, c := range []string{"pacman", "hyprctl", "kitty", "waybar", "hyprlock", "hypridle", "hyprpaper", "fuzzel", "mako", "nmcli", "bluetoothctl", "mmcli", "wpctl", "grim", "slurp", "wl-copy", "systemctl", "loginctl", "jq", "go", "brightnessctl", "playerctl", "powerprofilesctl", "sensors", "cava", "pfetch", "uname", "fc-match", "fc-cache"} {
		if r.Exists(c) {
			add("command: "+c, "pass", "found", "")
		} else {
			add("command: "+c, "warn", "missing", "Install the package that provides "+c)
		}
	}
	if r.Exists("fprintd-enroll") && r.Exists("fprintd-verify") {
		add("fingerprint tools", "pass", "fprintd commands found", "")
	} else {
		add("fingerprint tools", "warn", "fprintd not installed", "Install fprintd to use hyprglass touchid")
	}
	if r.Exists("fc-match") {
		checkFont := func(label, query string) {
			out, err := r.Run("fc-match", "-f", "%{family}\n", query)
			family := strings.TrimSpace(out)
			if err != nil || family == "" {
				add("font: "+label, "warn", "not detected", "Install ttf-jetbrains-mono-nerd and ttf-nerd-fonts-symbols-mono, then run hyprglass icons repair")
				return
			}
			lower := strings.ToLower(strings.ReplaceAll(family, " ", ""))
			queryNorm := strings.ToLower(strings.ReplaceAll(query, " ", ""))
			if strings.Contains(lower, queryNorm) || strings.Contains(lower, "symbolsnerdfont") || strings.Contains(lower, "jetbrainsmono") {
				add("font: "+label, "pass", family, "")
			} else {
				add("font: "+label, "warn", "resolved to "+family, "Run hyprglass icons repair")
			}
		}
		checkFont("JetBrains Mono Nerd", "JetBrainsMono Nerd Font")
		checkFont("Symbols Nerd Font Mono", "Symbols Nerd Font Mono")
	} else {
		add("fontconfig", "warn", "fc-match missing", "Install fontconfig, ttf-jetbrains-mono-nerd, and ttf-nerd-fonts-symbols-mono")
	}
	if os.Geteuid() == 0 {
		add("user level", "warn", "running as root", "Run user-level checks as normal user")
	} else {
		add("user level", "pass", "not root", "")
	}
	files := []requiredFile{
		{RepoPath: "config/hypr/hyprland.conf", InstalledPath: ".config/hypr/hyprland.conf"},
		{RepoPath: "config/hyprlock/hyprlock.conf", InstalledPath: ".config/hypr/hyprlock.conf"},
		{RepoPath: "config/hypridle/hypridle.conf", InstalledPath: ".config/hypr/hypridle.conf"},
		{RepoPath: "config/waybar/config.jsonc", InstalledPath: ".config/waybar/config.jsonc"},
		{RepoPath: "config/waybar/style.css", InstalledPath: ".config/waybar/style.css"},
		{RepoPath: "config/kitty/kitty.conf", InstalledPath: ".config/kitty/kitty.conf"},
		{RepoPath: "assets/wallpapers/hyprglass-dusk.png", InstalledPath: ".config/hypr/assets/wallpapers/hyprglass-dusk.png"},
	}
	for _, f := range files {
		if _, err := os.Stat(f.RepoPath); err == nil {
			add("repo file: "+f.RepoPath, "pass", "exists", "")
		} else {
			home, _ := os.UserHomeDir()
			alt := filepath.Join(home, f.InstalledPath)
			if _, e := os.Stat(alt); e == nil {
				add("installed file: "+f.RepoPath, "pass", "exists in ~/"+f.InstalledPath, "")
			} else {
				add("file: "+f.RepoPath, "warn", "not found in cwd or installed config", "Run from repo root or install configs")
			}
		}
	}
	status := "pass"
	for _, c := range cs {
		if c.Status == "fail" {
			status = "fail"
			break
		}
		if c.Status == "warn" && status == "pass" {
			status = "warn"
		}
	}
	return Result{Status: status, Checks: cs}
}
