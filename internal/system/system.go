package system

import (
	"bufio"
	"encoding/json"
	"fmt"
	"hyprglass/internal/command"
	"hyprglass/internal/platform"
	"hyprglass/internal/tui"
	"os"
	"os/exec"
	"strings"
)

type Status struct {
	Platform          platform.Info `json:"platform"`
	Kernel            string        `json:"kernel"`
	Pacman            bool          `json:"pacman"`
	CachyRateMirrors  bool          `json:"cachyRateMirrors"`
	CHWD              bool          `json:"chwd"`
	LinuxCachyos      string        `json:"linuxCachyos,omitempty"`
	RecommendedSource string        `json:"recommendedPackageSource"`
}

func RunTUI(r command.Runner, args []string) {
	if len(args) > 0 {
		switch args[0] {
		case "--json":
			b, _ := json.MarshalIndent(Collect(r), "", "  ")
			fmt.Println(string(b))
			return
		case "rate-mirrors":
			runCachyRateMirrors(r)
			return
		case "update":
			runSystemUpdate(r)
			return
		case "chwd-list":
			runCHWDList(r)
			return
		case "chwd-auto":
			runCHWDAuto(r)
			return
		}
	}
	menu(r)
}

func Collect(r command.Runner) Status {
	info := platform.Read()
	kernel := ""
	if r.Exists("uname") {
		out, err := r.Run("uname", "-r")
		if err == nil {
			kernel = strings.TrimSpace(out)
		}
	}
	linuxCachyos := ""
	if r.Exists("pacman") {
		out, err := r.Run("pacman", "-Q", "linux-cachyos")
		if err == nil {
			linuxCachyos = strings.TrimSpace(out)
		}
	}
	return Status{
		Platform:          info,
		Kernel:            kernel,
		Pacman:            r.Exists("pacman"),
		CachyRateMirrors:  r.Exists("cachyos-rate-mirrors"),
		CHWD:              r.Exists("chwd"),
		LinuxCachyos:      linuxCachyos,
		RecommendedSource: info.PackageFile(),
	}
}

func menu(r command.Runner) {
	reader := bufio.NewReader(os.Stdin)
	for {
		st := Collect(r)
		tui.Clear()
		tui.Header("System")
		fmt.Println("distro:", st.Platform.DisplayName())
		fmt.Println("arch-like:", yesNo(st.Platform.ArchLike))
		fmt.Println("cachyos:", yesNo(st.Platform.CachyOS))
		fmt.Println("kernel:", dash(st.Kernel))
		if st.LinuxCachyos != "" {
			fmt.Println("linux-cachyos:", st.LinuxCachyos)
		}
		fmt.Println("package source:", st.RecommendedSource)
		fmt.Println()
		fmt.Println("  1  Full system update")
		fmt.Println("  2  Rank CachyOS mirrors")
		fmt.Println("  3  List CachyOS hardware profiles")
		fmt.Println("  4  Run CachyOS hardware auto-configuration")
		fmt.Println("  q  Back")
		fmt.Print("\nSelect: ")
		switch strings.ToLower(tui.ReadLine(reader)) {
		case "1":
			runSystemUpdate(r)
			tui.Pause(reader)
		case "2":
			runCachyRateMirrors(r)
			tui.Pause(reader)
		case "3":
			runCHWDList(r)
			tui.Pause(reader)
		case "4":
			runCHWDAuto(r)
			tui.Pause(reader)
		case "q", "":
			return
		default:
			fmt.Println("Unknown selection.")
			tui.Pause(reader)
		}
	}
}

func runSystemUpdate(r command.Runner) {
	if !r.Exists("pacman") || !r.Exists("sudo") {
		fmt.Println("pacman and sudo are required for system updates.")
		return
	}
	if !confirm("Run sudo pacman -Syu now") {
		fmt.Println("Canceled.")
		return
	}
	runInteractive("sudo", "pacman", "-Syu")
}

func runCachyRateMirrors(r command.Runner) {
	if !r.Exists("cachyos-rate-mirrors") || !r.Exists("sudo") {
		fmt.Println("cachyos-rate-mirrors is not available. This action is only for CachyOS systems with the tool installed.")
		return
	}
	if !confirm("Run sudo cachyos-rate-mirrors now") {
		fmt.Println("Canceled.")
		return
	}
	runInteractive("sudo", "cachyos-rate-mirrors")
}

func runCHWDList(r command.Runner) {
	if !r.Exists("chwd") {
		fmt.Println("chwd is not installed. On CachyOS this normally comes from CachyOS hardware detection tooling.")
		return
	}
	out, err := r.Run("chwd", "--list-all")
	if err != nil {
		fmt.Println("chwd failed:", err)
	}
	fmt.Print(out)
}

func runCHWDAuto(r command.Runner) {
	if !r.Exists("chwd") || !r.Exists("sudo") {
		fmt.Println("chwd and sudo are required.")
		return
	}
	fmt.Println("This asks CachyOS hardware detection to install/configure driver profiles for this machine.")
	fmt.Println("Use it on CachyOS, not as a generic Arch conversion button.")
	if !confirm("Run sudo chwd -a now") {
		fmt.Println("Canceled.")
		return
	}
	runInteractive("sudo", "chwd", "-a")
}

func runInteractive(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Println("Command failed:", err)
	}
}

func confirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s? Type yes: ", prompt)
	answer, _ := reader.ReadString('\n')
	return strings.TrimSpace(strings.ToLower(answer)) == "yes"
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func dash(v string) string {
	if strings.TrimSpace(v) == "" {
		return "-"
	}
	return v
}
