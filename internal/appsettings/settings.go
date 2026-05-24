package appsettings

import (
	"bufio"
	"encoding/json"
	"fmt"
	"hyprglass/internal/audio"
	"hyprglass/internal/bluetooth"
	"hyprglass/internal/command"
	"hyprglass/internal/display"
	"hyprglass/internal/doctor"
	"hyprglass/internal/icons"
	"hyprglass/internal/laptop"
	"hyprglass/internal/lte"
	"hyprglass/internal/prefs"
	hgsystem "hyprglass/internal/system"
	"hyprglass/internal/tui"
	"hyprglass/internal/wifi"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func Run(r command.Runner, args []string, version string) {
	if len(args) > 0 {
		switch args[0] {
		case "apply":
			applyOnly(r, contains(args[1:], "--no-reload"), contains(args[1:], "--with-display") || contains(args[1:], "--display"))
			return
		case "defaults":
			p := prefs.Default()
			if err := prefs.Save(p); err != nil {
				fmt.Println("Could not write defaults:", err)
				return
			}
			if err := prefs.ApplyAll(p); err != nil {
				fmt.Println("Could not apply defaults:", err)
				return
			}
			fmt.Println("Default Hyprglass preferences written.")
			return
		case "--json":
			b, _ := json.MarshalIndent(prefs.Load(), "", "  ")
			fmt.Println(string(b))
			return
		}
	}
	runMenu(r, version)
}

func runMenu(r command.Runner, version string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		p := prefs.Load()
		clear()
		tui.Header("Settings")
		fmt.Println("Hyprglass", version)
		fmt.Printf("appearance: %s, %s accent\n", p.ThemeMode, p.Accent)
		fmt.Printf("display: preferred resolution, scale %s\n", p.MonitorScale)
		fmt.Printf("keyboard: %s %s\n", p.KeyboardLayout, emptyDash(p.KeyboardVariant))
		fmt.Println()
		fmt.Println("  1  Appearance")
		fmt.Println("  2  Display, scaling, and keyboard")
		fmt.Println("  3  Network, Bluetooth, and modem")
		fmt.Println("  4  Audio")
		fmt.Println("  5  Power and battery")
		fmt.Println("  6  Update")
		fmt.Println("  7  Developer options")
		fmt.Println("  q  Close")
		fmt.Print("\nSelect: ")
		choice := readLine(reader)
		switch strings.ToLower(choice) {
		case "1":
			appearance(reader, r)
		case "2":
			displayKeyboard(reader, r)
		case "3":
			connectivity(reader, r)
		case "4":
			audio.RunTUI(r)
			pause(reader)
		case "5":
			laptop.RunTUI(r, nil)
		case "6", "u":
			update(reader, r)
		case "7", "d":
			developerOptions(reader, r)
		case "q", "":
			fmt.Println("Closed.")
			return
		default:
			fmt.Println("Unknown selection.")
			pause(reader)
		}
	}
}

func appearance(reader *bufio.Reader, r command.Runner) {
	p := prefs.Load()
	clear()
	tui.Header("Appearance")
	fmt.Println("Current:", p.ThemeMode, p.Accent)
	fmt.Print("Theme mode [dark/light] (enter keeps current): ")
	mode := readLine(reader)
	if mode != "" {
		p.ThemeMode = mode
	}
	fmt.Println("Accent colors:", strings.Join(prefs.AccentNames(), ", "))
	fmt.Print("Accent (enter keeps current): ")
	accent := readLine(reader)
	if accent != "" {
		p.Accent = accent
	}
	saveApplyReload(reader, r, p, prefs.ApplyVisuals)
}

func displayAndScaling(reader *bufio.Reader, r command.Runner) {
	p := prefs.Load()
	clear()
	tui.Header("Display and scaling")
	if r.Exists("hyprctl") {
		display.RunTUI(r)
		fmt.Println()
	}
	fmt.Println("Recommended laptop scale choices: auto, 1.25, 1.5, 1.75, 2")
	fmt.Println("Use auto first. Use 1.75/2 on 4K 14-16 inch panels.")
	fmt.Print("Scale (enter keeps current): ")
	scale := readLine(reader)
	if scale != "" {
		p.MonitorScale = scale
	}
	saveApplyReload(reader, r, p, prefs.ApplyDisplay)
}

func displayKeyboard(reader *bufio.Reader, r command.Runner) {
	p := prefs.Load()
	clear()
	tui.Header("Display and keyboard")
	if r.Exists("hyprctl") {
		display.RunTUI(r)
		fmt.Println()
	}
	fmt.Println("Scale choices: auto, 1.25, 1.5, 1.75, 2")
	fmt.Print("Display scale (enter keeps current): ")
	scale := readLine(reader)
	if scale != "" {
		p.MonitorScale = scale
	}
	fmt.Println()
	fmt.Println("Keyboard examples: us, it, es, gb, de. Variant can stay empty.")
	fmt.Print("Keyboard layout (enter keeps current): ")
	layout := readLine(reader)
	if layout != "" {
		p.KeyboardLayout = layout
	}
	fmt.Print("Keyboard variant (enter keeps current): ")
	variant := readLine(reader)
	if variant != "" {
		p.KeyboardVariant = variant
	}
	saveApplyReload(reader, r, p, prefs.ApplyDisplayAndInput)
}

func keyboard(reader *bufio.Reader, r command.Runner) {
	p := prefs.Load()
	clear()
	tui.Header("Keyboard")
	fmt.Println("Examples: us, it, es, gb, de. Variant can stay empty.")
	fmt.Print("Layout (enter keeps current): ")
	layout := readLine(reader)
	if layout != "" {
		p.KeyboardLayout = layout
	}
	fmt.Print("Variant (enter for empty/current): ")
	variant := readLine(reader)
	if variant != "" {
		p.KeyboardVariant = variant
	}
	saveApplyReload(reader, r, p, prefs.ApplyInput)
}

func wallpaperRepair(reader *bufio.Reader, r command.Runner) {
	clear()
	tui.Header("Wallpaper repair")
	if err := installWallpaperFromSource(); err != nil {
		fmt.Println("Wallpaper repair failed:", err)
		pause(reader)
		return
	}
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		if r.Exists("pkill") {
			_, _ = r.Run("pkill", "-x", "hyprpaper")
		}
		startDetached("hyprpaper")
		if r.Exists("hyprctl") {
			wallpaper := filepath.Join(homeDir(), ".config", "hypr", "assets", "wallpapers", "hyprglass-dusk.png")
			_, _ = r.Run("hyprctl", "hyprpaper", "wallpaper", ", "+wallpaper+", cover")
		}
	}
	fmt.Println("Wallpaper asset copied and hyprpaper.conf rewritten using current hyprpaper syntax.")
	pause(reader)
}

func connectivity(reader *bufio.Reader, r command.Runner) {
	for {
		clear()
		tui.Header("Network, Bluetooth, and modem")
		fmt.Println("  1  Wi-Fi")
		fmt.Println("  2  Bluetooth")
		fmt.Println("  3  Modem status")
		fmt.Println("  4  Configure modem autounlock/autoconnect")
		fmt.Println("  q  Back")
		fmt.Print("\nSelect: ")
		switch strings.ToLower(readLine(reader)) {
		case "1":
			wifi.RunTUI(r)
			pause(reader)
		case "2":
			bluetooth.RunTUI(r)
			pause(reader)
		case "3":
			lte.RunTUI(r)
			pause(reader)
		case "4":
			modemAutounlock(reader, r)
		case "q", "":
			return
		}
	}
}

func developerOptions(reader *bufio.Reader, r command.Runner) {
	for {
		clear()
		tui.Header("Developer options")
		fmt.Println("Advanced actions are hidden here so the normal Settings screen stays simple.")
		fmt.Println()
		fmt.Println("  1  Doctor")
		fmt.Println("  2  Services")
		fmt.Println("  3  Wallpaper repair")
		fmt.Println("  4  Icon/font repair")
		fmt.Println("  5  System and CachyOS")
		fmt.Println("  6  Raw keyboard page")
		fmt.Println("  7  Raw display page")
		fmt.Println("  q  Back")
		fmt.Print("\nSelect: ")
		switch strings.ToLower(readLine(reader)) {
		case "1":
			tui.PrintChecks(doctor.Run(r))
			pause(reader)
		case "2":
			services(reader, r)
		case "3":
			wallpaperRepair(reader, r)
		case "4":
			icons.RunTUI(r)
		case "5":
			hgsystem.RunTUI(r, nil)
		case "6":
			keyboard(reader, r)
		case "7":
			displayAndScaling(reader, r)
		case "q", "":
			return
		default:
			fmt.Println("Unknown selection.")
			pause(reader)
		}
	}
}

func modemAutounlock(reader *bufio.Reader, r command.Runner) {
	p := prefs.Load()
	clear()
	tui.Header("Modem autounlock")
	fmt.Println("This creates a root-owned systemd service and stores the SIM PIN in /etc/hyprglass/modem.env with 0600 permissions.")
	fmt.Print("APN (enter keeps current): ")
	apn := readLine(reader)
	if apn != "" {
		p.ModemAPN = apn
	}
	fmt.Print("SIM PIN (empty disables PIN storage): ")
	pin := readLine(reader)
	args := []string{filepath.Join(sourceRoot(), "scripts", "hyprglass-modem-autounlock-install.sh")}
	if p.ModemAPN != "" {
		args = append(args, "--apn", p.ModemAPN)
	}
	if pin != "" {
		args = append(args, "--pin", pin)
		p.ModemPINSet = true
	} else {
		p.ModemPINSet = false
	}
	if !r.Exists("sudo") {
		fmt.Println("sudo is required to install the modem service.")
		pause(reader)
		return
	}
	if _, err := os.Stat(args[0]); err != nil {
		fmt.Println("Installer script not found:", args[0])
		pause(reader)
		return
	}
	cmd := exec.Command("sudo", append([]string{"bash"}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Println("Modem service install failed:", err)
		pause(reader)
		return
	}
	_ = prefs.Save(p)
	fmt.Println("Modem service installed.")
	pause(reader)
}

func services(reader *bufio.Reader, r command.Runner) {
	clear()
	tui.Header("Services")
	services := []string{"NetworkManager.service", "bluetooth.service", "ModemManager.service", "power-profiles-daemon.service"}
	if r.Exists("systemctl") {
		for _, svc := range services {
			out, _ := r.Run("systemctl", "is-enabled", svc)
			fmt.Printf("%-34s %s", svc, out)
		}
	}
	fmt.Print("\nEnable and start recommended services with sudo? [y/N] ")
	if strings.EqualFold(readLine(reader), "y") {
		for _, svc := range services {
			cmd := exec.Command("sudo", "systemctl", "enable", "--now", svc)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Run()
		}
	}
	pause(reader)
}

func update(reader *bufio.Reader, r command.Runner) {
	clear()
	tui.Header("Update")
	root := sourceRoot()
	if root == "" {
		fmt.Println("Source checkout not found. Download the repo zip again and run ./install.sh --update.")
		pause(reader)
		return
	}
	script := filepath.Join(root, "install.sh")
	if _, err := os.Stat(script); err != nil {
		fmt.Println("install.sh is missing in", root)
		pause(reader)
		return
	}
	fmt.Print("Run Hyprglass update now? [y/N] ")
	if !strings.EqualFold(readLine(reader), "y") {
		return
	}
	cmd := exec.Command("bash", script, "--update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	_ = cmd.Run()
	pause(reader)
}

func saveApplyReload(reader *bufio.Reader, r command.Runner, p prefs.Preferences, apply func(prefs.Preferences) error) {
	if err := prefs.Save(p); err != nil {
		fmt.Println("Could not save preferences:", err)
		pause(reader)
		return
	}
	if err := apply(p); err != nil {
		fmt.Println("Could not apply preferences:", err)
		pause(reader)
		return
	}
	applyGSettings(r, p.ThemeMode)
	reloadSession(r)
	fmt.Println("Applied.")
	pause(reader)
}

func applyOnly(r command.Runner, noReload bool, withDisplay bool) {
	p := prefs.Load()
	if err := prefs.Save(p); err != nil {
		fmt.Println("Could not save preferences:", err)
		return
	}
	apply := prefs.Apply
	if withDisplay {
		apply = prefs.ApplyAll
	}
	if err := apply(p); err != nil {
		fmt.Println("Could not apply preferences:", err)
		return
	}
	applyGSettings(r, p.ThemeMode)
	if !noReload {
		reloadSession(r)
	}
	fmt.Println("Hyprglass preferences applied.")
}

func applyGSettings(r command.Runner, mode string) {
	if !r.Exists("gsettings") {
		return
	}
	value := "prefer-dark"
	if mode == "light" {
		value = "prefer-light"
	}
	_, _ = r.Run("gsettings", "set", "org.gnome.desktop.interface", "color-scheme", value)
}

func reloadSession(r command.Runner) {
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") == "" {
		return
	}
	if r.Exists("hyprctl") {
		_, _ = r.Run("hyprctl", "reload")
	}
	// Waybar supports a reload action on signal. Avoid killing it from inside Settings:
	// children spawned from a terminal can inherit the terminal session and die when Settings closes.
	if r.Exists("pkill") {
		_, _ = r.Run("pkill", "-SIGUSR2", "-x", "waybar")
	}
	ensureDetachedProcess(r, "waybar")
	ensureDetachedProcess(r, "mako")
}

func installWallpaperFromSource() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	root := sourceRoot()
	if root == "" {
		return fmt.Errorf("source checkout not found")
	}
	source := filepath.Join(root, "assets", "wallpapers", "hyprglass-dusk.png")
	target := filepath.Join(home, ".config", "hypr", "assets", "wallpapers", "hyprglass-dusk.png")
	if err := copyFile(source, target); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(home, ".config", "hypr", "hyprpaper.conf"), []byte(fmt.Sprintf(`# Hyprglass hyprpaper configuration
# Current hyprpaper syntax: one fallback wallpaper block for every monitor.
wallpaper {
    monitor =
    path = %s
    fit_mode = cover
}

splash = false
ipc = true
`, target)), 0o644)
}

func sourceRoot() string {
	candidates := []string{os.Getenv("HYPRGLASS_SOURCE_ROOT")}
	if home, err := os.UserHomeDir(); err == nil {
		if b, err := os.ReadFile(filepath.Join(home, ".config", "hyprglass", "source-root")); err == nil {
			candidates = append(candidates, strings.TrimSpace(string(b)))
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		for {
			candidates = append(candidates, cwd)
			parent := filepath.Dir(cwd)
			if parent == cwd {
				break
			}
			cwd = parent
		}
	}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(c, "install.sh")); err == nil {
			if _, err := os.Stat(filepath.Join(c, "cmd", "hyprglass", "main.go")); err == nil {
				return c
			}
		}
	}
	return ""
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}

func ensureDetachedProcess(r command.Runner, name string) {
	if !r.Exists("pgrep") || !r.Exists(name) {
		if r.Exists(name) {
			startDetached(name)
		}
		return
	}
	if _, err := r.Run("pgrep", "-x", name); err != nil {
		startDetached(name)
	}
}

func startDetached(name string) {
	if _, err := exec.LookPath(name); err != nil {
		return
	}
	devNull, _ := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if devNull != nil {
		defer devNull.Close()
	}
	cmd := exec.Command(name)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if devNull != nil {
		cmd.Stdin = devNull
		cmd.Stdout = devNull
		cmd.Stderr = devNull
	}
	if err := cmd.Start(); err == nil {
		_ = cmd.Process.Release()
	}
}

func clear() {
	if os.Getenv("TERM") != "" {
		fmt.Print("\033[H\033[2J")
	}
}

func readLine(r *bufio.Reader) string {
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

func pause(r *bufio.Reader) {
	fmt.Print("\nPress Enter to continue.")
	_, _ = r.ReadString('\n')
}

func emptyDash(v string) string {
	if strings.TrimSpace(v) == "" {
		return "-"
	}
	return v
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
