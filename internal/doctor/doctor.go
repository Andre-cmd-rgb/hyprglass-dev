package doctor

import (
	"hyprglass/internal/command"
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
	if b, err := os.ReadFile("/etc/os-release"); err == nil && strings.Contains(strings.ToLower(string(b)), "arch") {
		add("arch detection", "pass", "/etc/os-release looks like Arch", "")
	} else {
		add("arch detection", "warn", "this environment does not look like Arch", "Run on Arch for package verification")
	}
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		add("hyprland session", "pass", "Hyprland session variable present", "")
	} else {
		add("hyprland session", "warn", "not inside Hyprland; runtime compositor checks skipped", "Start Hyprland then rerun doctor")
	}
	for _, c := range []string{"hyprctl", "kitty", "waybar", "hyprlock", "hypridle", "hyprpaper", "fuzzel", "mako", "nmcli", "bluetoothctl", "mmcli", "wpctl", "grim", "slurp", "wl-copy", "systemctl", "loginctl", "jq", "go", "brightnessctl", "playerctl"} {
		if r.Exists(c) {
			add("command: "+c, "pass", "found", "")
		} else {
			add("command: "+c, "warn", "missing", "Install the package that provides "+c)
		}
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
