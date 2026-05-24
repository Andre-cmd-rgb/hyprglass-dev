package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"hyprglass/internal/appsettings"
	"hyprglass/internal/audio"
	"hyprglass/internal/bluetooth"
	"hyprglass/internal/command"
	"hyprglass/internal/display"
	"hyprglass/internal/doctor"
	"hyprglass/internal/fileutil"
	"hyprglass/internal/icons"
	"hyprglass/internal/laptop"
	"hyprglass/internal/lte"
	"hyprglass/internal/srcroot"
	hgsystem "hyprglass/internal/system"
	"hyprglass/internal/tui"
	"hyprglass/internal/wifi"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"
)

var (
	version    = "dev"
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
	case "info":
		info(r)
	case "icons":
		icons.Run(r, args[1:])
	case "doctor":
		res := doctor.Run(r)
		if slices.Contains(args[1:], "--json") {
			b, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(b))
			return
		}
		tui.PrintChecks(res)
	case "wifi":
		wifi.RunTUI(r)
	case "bluetooth":
		if slices.Contains(args[1:], "--status") {
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
	case "laptop":
		laptop.RunTUI(r, args[1:])
	case "settings":
		appsettings.Run(r, args[1:], version, sourceRoot)
	case "system", "cachyos":
		hgsystem.RunTUI(r, args[1:])
	case "power":
		power(r)
	case "visualizer", "cava":
		visualizer(r)
	case "touchid", "fingerprint":
		touchID(r, args[1:])
	case "update":
		update()
	case "repair":
		repair()
	case "wallpaper":
		wallpaper(r, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		help()
		os.Exit(2)
	}
}

func help() {
	fmt.Print(`Hyprglass is a fast, glassy Wayland desktop built on Hyprland for focused work.

Usage:
  hyprglass --help
  hyprglass version
  hyprglass info
  hyprglass doctor [--json]
  hyprglass icons [status|repair]
  hyprglass wifi | bluetooth | lte | audio | display | laptop | settings | system | power
  hyprglass cava
  hyprglass touchid [status|enroll|verify]
  hyprglass update
  hyprglass repair
  hyprglass wallpaper [apply|generate]
`)
}

func power(r command.Runner) {
	type action struct {
		Key         string
		Label       string
		Command     string
		Args        []string
		Destructive bool
	}
	actions := []action{
		{Key: "1", Label: "Lock session", Command: "loginctl", Args: []string{"lock-session"}},
		{Key: "2", Label: "Suspend", Command: "systemctl", Args: []string{"suspend"}},
		{Key: "3", Label: "Log out of Hyprland", Command: "hyprctl", Args: []string{"dispatch", "exit"}, Destructive: true},
		{Key: "4", Label: "Reboot", Command: "systemctl", Args: []string{"reboot"}, Destructive: true},
		{Key: "5", Label: "Power off", Command: "systemctl", Args: []string{"poweroff"}, Destructive: true},
		{Key: "q", Label: "Cancel"},
	}

	reader := bufio.NewReader(os.Stdin)
	tui.Header("Power")
	for _, a := range actions {
		fmt.Printf("  %s  %s\n", a.Key, a.Label)
	}
	fmt.Print("\nSelect: ")
	choice := strings.TrimSpace(strings.ToLower(tui.ReadLine(reader)))
	for _, a := range actions {
		if choice != a.Key {
			continue
		}
		if a.Command == "" {
			fmt.Println("Canceled.")
			return
		}
		if !r.Exists(a.Command) {
			fmt.Println(a.Command, "is not installed or not on PATH")
			return
		}
		if a.Destructive {
			fmt.Printf("Type yes to confirm %q: ", a.Label)
			confirm := strings.TrimSpace(strings.ToLower(tui.ReadLine(reader)))
			if confirm != "yes" {
				fmt.Println("Canceled.")
				return
			}
		}
		out, err := r.Run(a.Command, a.Args...)
		if err != nil {
			fmt.Println("Action failed:", err)
			fmt.Print(out)
			return
		}
		fmt.Print(out)
		return
	}
	fmt.Println("Unknown selection.")
}

func info(r command.Runner) {
	if r.Exists("pfetch") {
		out, err := r.Run("pfetch")
		fmt.Print(filterDconfWarnings(out))
		if err != nil {
			fmt.Println("pfetch failed:", err)
		}
		return
	}
	if r.Exists("fastfetch") {
		out, err := r.Run("fastfetch")
		fmt.Print(filterDconfWarnings(out))
		if err != nil {
			fmt.Println("fastfetch failed:", err)
		}
		return
	}
	fmt.Println("pfetch is missing. Install pfetch-rs.")
}

func visualizer(r command.Runner) {
	if !r.Exists("cava") {
		fmt.Println("cava is missing. Install cava.")
		return
	}
	if err := syscall.Exec("cava", []string{"cava"}, os.Environ()); err != nil {
		fmt.Println("could not start cava:", err)
	}
}

func touchID(r command.Runner, args []string) {
	mode := "status"
	if len(args) > 0 {
		mode = args[0]
	}
	tui.Header("Touch ID / Fingerprint")
	if !r.Exists("fprintd-enroll") || !r.Exists("fprintd-verify") {
		fmt.Println("Fingerprint tools are missing. Install fprintd.")
		fmt.Println("PAM is not edited automatically; review distro docs before enabling fingerprint auth for sudo/login.")
		return
	}
	switch mode {
	case "status":
		if r.Exists("fprintd-list") {
			user := os.Getenv("USER")
			var out string
			var err error
			if user == "" {
				out, err = r.Run("fprintd-list")
			} else {
				out, err = r.Run("fprintd-list", user)
			}
			if err != nil {
				fmt.Println("Could not list enrolled fingerprints. Make sure fprintd can access the system bus and a fingerprint reader is present.")
			} else if strings.TrimSpace(out) != "" {
				fmt.Print(strings.TrimSpace(out))
				fmt.Println()
			}
		}
		if r.Exists("systemctl") {
			out, err := r.Run("systemctl", "status", "fprintd.service", "--no-pager")
			if err != nil && strings.Contains(out, "Failed to connect") {
				fmt.Println("fprintd service status unavailable from this session.")
			} else {
				for _, line := range firstLines(out, 8) {
					fmt.Println(line)
				}
			}
		}
		fmt.Println("Commands: hyprglass touchid enroll | hyprglass touchid verify")
	case "enroll":
		out, err := r.Run("fprintd-enroll")
		fmt.Print(out)
		if err != nil {
			fmt.Println("Enrollment failed:", err)
			os.Exit(1)
		}
	case "verify":
		out, err := r.Run("fprintd-verify")
		fmt.Print(out)
		if err != nil {
			fmt.Println("Verification failed:", err)
			os.Exit(1)
		}
	default:
		fmt.Println("Usage: hyprglass touchid [status|enroll|verify]")
	}
}

func wallpaper(r command.Runner, args []string) {
	mode := "apply"
	if len(args) > 0 {
		mode = args[0]
	}
	root := srcroot.Find(sourceRoot)
	if root == "" {
		fmt.Println("wallpaper commands require a Hyprglass source checkout")
		os.Exit(1)
	}
	source := filepath.Join(root, "assets", "wallpapers", "hyprglass-dusk.png")
	switch mode {
	case "generate":
		script := filepath.Join(root, "scripts", "generate-wallpaper.py")
		out, err := r.Run("python3", script)
		if err != nil {
			fmt.Println("wallpaper generation failed:", err)
			fmt.Print(out)
			os.Exit(1)
		}
		fmt.Print(out)
	case "apply":
	default:
		fmt.Println("Usage: hyprglass wallpaper [apply|generate]")
		os.Exit(2)
	}
	installed, err := installWallpaper(source)
	if err != nil {
		fmt.Println("could not install wallpaper:", err)
		os.Exit(1)
	}
	restartHyprpaper(r, installed)
	fmt.Println("Wallpaper installed to", installed)
}

func update() {
	root := srcroot.Find(sourceRoot)
	if root == "" {
		fmt.Println("Hyprglass could not find its source checkout, so it cannot auto-update. Run: sudo pacman -Syu, then rerun hyprglass doctor")
		return
	}
	script := filepath.Join(root, "install.sh")
	if _, err := os.Stat(script); err != nil {
		fmt.Println("Hyprglass could not find install.sh in its source checkout, so it cannot auto-update. Run: sudo pacman -Syu, then rerun hyprglass doctor")
		return
	}
	ensureExecutableBits(root)
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
	r := command.RealRunner{}
	fmt.Println("Hyprglass repair: refreshing generated config, wallpaper, and session surfaces.")
	root := srcroot.Find(sourceRoot)
	wallpaperPath := ""
	if root != "" {
		source := filepath.Join(root, "assets", "wallpapers", "hyprglass-dusk.png")
		installed, err := installWallpaper(source)
		if err != nil {
			fmt.Println("Wallpaper repair failed:", err)
		} else {
			wallpaperPath = installed
			fmt.Println("Wallpaper config repaired.")
		}
	} else {
		fmt.Println("Source checkout not found; wallpaper file was not recopied.")
	}
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") == "" {
		fmt.Println("Not inside Hyprland; session components were not restarted.")
		return
	}
	if wallpaperPath != "" {
		restartHyprpaper(r, wallpaperPath)
	}
	if r.Exists("pkill") {
		_, _ = r.Run("pkill", "-x", "waybar")
		_, _ = r.Run("pkill", "-x", "mako")
		_, _ = r.Run("pkill", "-x", "hypridle")
	}
	tui.Launch("waybar")
	tui.Launch("mako")
	tui.Launch("hypridle")
	fmt.Println("Session surfaces restarted. Network services were not touched.")
}

func ensureExecutableBits(root string) {
	_ = os.Chmod(filepath.Join(root, "install.sh"), 0o755)
	_ = os.Chmod(filepath.Join(root, "uninstall.sh"), 0o755)
	for _, pattern := range []string{"scripts/*.sh", "scripts/*.py"} {
		matches, _ := filepath.Glob(filepath.Join(root, pattern))
		for _, m := range matches {
			_ = os.Chmod(m, 0o755)
		}
	}
}

func installWallpaper(source string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	targetDir := filepath.Join(home, ".config", "hypr", "assets", "wallpapers")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", err
	}
	target := filepath.Join(targetDir, "hyprglass-dusk.png")
	if err := fileutil.Copy(source, target); err != nil {
		return "", err
	}
	return target, writeHyprpaperConfig(filepath.Join(home, ".config", "hypr", "hyprpaper.conf"), target)
}

func writeHyprpaperConfig(path, wallpaper string) error {
	data := fmt.Sprintf(`# Hyprglass hyprpaper configuration
# Current hyprpaper syntax: one fallback wallpaper block for every monitor.
wallpaper {
    monitor =
    path = %s
    fit_mode = cover
}

splash = false
ipc = true
`, wallpaper)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(data), 0o644)
}

func restartHyprpaper(r command.Runner, wallpaper string) {
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") == "" {
		return
	}
	if r.Exists("pkill") {
		_, _ = r.Run("pkill", "-x", "hyprpaper")
	}
	if r.Exists("hyprpaper") {
		cmd := exec.Command("hyprpaper")
		if err := cmd.Start(); err != nil {
			fmt.Println("Could not restart hyprpaper:", err)
		} else {
			_ = cmd.Process.Release()
		}
	}
	if wallpaper != "" && r.Exists("hyprctl") {
		time.Sleep(300 * time.Millisecond)
		_, _ = r.Run("hyprctl", "hyprpaper", "wallpaper", ", "+wallpaper+", cover")
	}
}

func firstLines(s string, limit int) []string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		lines = append(lines, line)
		if len(lines) >= limit {
			break
		}
	}
	return lines
}

func filterDconfWarnings(out string) string {
	var kept []string
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "dconf-CRITICAL") ||
			strings.Contains(line, "unable to create file '/run/user") ||
			strings.Contains(line, "dconf will not work properly") {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}
