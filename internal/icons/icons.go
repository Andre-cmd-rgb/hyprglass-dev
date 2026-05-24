package icons

import (
	"bufio"
	"fmt"
	"hyprglass/internal/command"
	"hyprglass/internal/tui"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

var requiredFonts = []string{
	"JetBrainsMono Nerd Font",
	"JetBrainsMonoNL Nerd Font",
	"Symbols Nerd Font Mono",
	"Symbols Nerd Font",
}

var requiredPackages = []string{
	"fontconfig",
	"ttf-jetbrains-mono-nerd",
	"ttf-nerd-fonts-symbols-mono",
}

func Run(r command.Runner, args []string) {
	mode := "status"
	if len(args) > 0 {
		mode = strings.ToLower(strings.TrimSpace(args[0]))
	}
	switch mode {
	case "repair", "fix":
		Repair(r)
	case "status", "doctor":
		Status(r)
	default:
		fmt.Println("usage: hyprglass icons [status|repair]")
	}
}

func RunTUI(r command.Runner) {
	reader := bufio.NewReader(os.Stdin)
	for {
		tui.Clear()
		tui.Header("Icon and font repair")
		fmt.Println("Waybar icons require Nerd Font text plus the Symbols Nerd Font fallback.")
		fmt.Println()
		printStatus(r)
		fmt.Println()
		fmt.Println("  1  Install missing icon fonts and rebuild font cache")
		fmt.Println("  2  Rebuild font cache only")
		fmt.Println("  3  Restart/reload Waybar")
		fmt.Println("  q  Back")
		fmt.Print("\nSelect: ")
		choice, _ := reader.ReadString('\n')
		switch strings.ToLower(strings.TrimSpace(choice)) {
		case "1":
			Repair(r)
			tui.Pause(reader)
		case "2":
			RefreshCache(r)
			tui.Pause(reader)
		case "3":
			ReloadWaybar(r)
			tui.Pause(reader)
		case "q", "":
			return
		default:
			fmt.Println("Unknown selection.")
			tui.Pause(reader)
		}
	}
}

func Status(r command.Runner) {
	tui.Header("Icon fonts")
	printStatus(r)
}

func Repair(r command.Runner) {
	tui.Header("Icon/font repair")
	if r.Exists("pacman") {
		missing := missingPackages(r)
		if len(missing) > 0 {
			fmt.Println("Installing:", strings.Join(missing, " "))
			cmd := exec.Command("sudo", append([]string{"pacman", "-S", "--needed"}, missing...)...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err != nil {
				fmt.Println("Font package install failed:", err)
				return
			}
		} else {
			fmt.Println("Required font packages already installed.")
		}
	} else {
		fmt.Println("pacman not found. Install these manually:", strings.Join(requiredPackages, " "))
	}
	RefreshCache(r)
	ReloadWaybar(r)
	printStatus(r)
}

func RefreshCache(r command.Runner) {
	if !r.Exists("fc-cache") {
		fmt.Println("fc-cache missing. Install fontconfig.")
		return
	}
	out, err := r.Run("fc-cache", "-f")
	if err != nil {
		fmt.Println("fc-cache failed:", err)
		if strings.TrimSpace(out) != "" {
			fmt.Println(strings.TrimSpace(out))
		}
		return
	}
	fmt.Println("Font cache rebuilt.")
}

func ReloadWaybar(r command.Runner) {
	if r.Exists("pkill") {
		_, _ = r.Run("pkill", "-SIGUSR2", "-x", "waybar")
	}
	if r.Exists("pgrep") {
		if _, err := r.Run("pgrep", "-x", "waybar"); err == nil {
			fmt.Println("Waybar reload signal sent.")
			return
		}
	}
	if _, err := exec.LookPath("waybar"); err != nil {
		fmt.Println("waybar is missing.")
		return
	}
	devNull, _ := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if devNull != nil {
		defer devNull.Close()
	}
	cmd := exec.Command("waybar")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if devNull != nil {
		cmd.Stdin = devNull
		cmd.Stdout = devNull
		cmd.Stderr = devNull
	}
	if err := cmd.Start(); err != nil {
		fmt.Println("Could not start Waybar:", err)
		return
	}
	_ = cmd.Process.Release()
	fmt.Println("Waybar started.")
}

func missingPackages(r command.Runner) []string {
	var missing []string
	for _, pkg := range requiredPackages {
		if _, err := r.Run("pacman", "-Q", pkg); err != nil {
			missing = append(missing, pkg)
		}
	}
	return missing
}

func printStatus(r command.Runner) {
	if !r.Exists("fc-match") {
		fmt.Println("fontconfig is missing: fc-match unavailable.")
		fmt.Println("Install:", strings.Join(requiredPackages, " "))
		return
	}
	for _, font := range requiredFonts {
		out, err := r.Run("fc-match", "-f", "%{family}\n", font)
		line := strings.TrimSpace(out)
		if err != nil || line == "" {
			fmt.Printf("%-30s missing\n", font)
			continue
		}
		if strings.Contains(strings.ToLower(line), strings.ToLower(font)) || familyContainsAny(line, []string{"jetbrainsmono", "symbols nerd font"}) {
			fmt.Printf("%-30s %s\n", font, line)
		} else {
			fmt.Printf("%-30s fallback: %s\n", font, line)
		}
	}
}

func familyContainsAny(line string, needles []string) bool {
	lower := strings.ToLower(strings.ReplaceAll(line, " ", ""))
	for _, n := range needles {
		if strings.Contains(lower, strings.ToLower(strings.ReplaceAll(n, " ", ""))) {
			return true
		}
	}
	return false
}
