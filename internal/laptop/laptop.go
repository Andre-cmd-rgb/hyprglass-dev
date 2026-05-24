package laptop

import (
	"bufio"
	"encoding/json"
	"fmt"
	"hyprglass/internal/command"
	"hyprglass/internal/tui"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type Battery struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Capacity int    `json:"capacity"`
}

type ThermalReading struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type Status struct {
	Batteries    []Battery        `json:"batteries"`
	PowerProfile string           `json:"powerProfile,omitempty"`
	Profiles     []string         `json:"profiles,omitempty"`
	Thermals     []ThermalReading `json:"thermals,omitempty"`
	Fingerprint  string           `json:"fingerprint"`
	LTE          string           `json:"lte"`
	Sleep        string           `json:"sleep"`
}

func RunTUI(r command.Runner, args []string) {
	status := Collect(r)
	if len(args) > 0 {
		switch args[0] {
		case "--json":
			b, _ := json.MarshalIndent(status, "", "  ")
			fmt.Println(string(b))
			return
		case "profile":
			setProfile(r, args[1:], status.Profiles)
			return
		case "suspend":
			runConfirmed(r, "Suspend", "systemctl", "suspend")
			return
		case "lock":
			runCommand(r, "loginctl", "lock-session")
			return
		case "hibernate":
			runConfirmed(r, "Hibernate", "systemctl", "hibernate")
			return
		}
	}
	printStatus(status)
	printMenu(r, status)
}

func Collect(r command.Runner) Status {
	return Status{
		Batteries:    batteries(),
		PowerProfile: currentProfile(r),
		Profiles:     profiles(r),
		Thermals:     thermals(r),
		Fingerprint:  fingerprintStatus(r),
		LTE:          lteStatus(r),
		Sleep:        sleepStatus(r),
	}
}

func printStatus(s Status) {
	tui.Header("Laptop")
	if len(s.Batteries) == 0 {
		fmt.Println("battery: no battery detected")
	} else {
		for _, b := range s.Batteries {
			fmt.Printf("battery: %-8s %3d%% %s\n", b.Name, b.Capacity, b.Status)
		}
	}
	if s.PowerProfile != "" {
		fmt.Println("power profile:", s.PowerProfile)
	} else {
		fmt.Println("power profile: unavailable")
	}
	if len(s.Thermals) == 0 {
		fmt.Println("thermal/fans: unavailable")
	} else {
		for _, t := range s.Thermals {
			if t.Unit == "RPM" {
				fmt.Printf("fan: %-24s %.0f RPM\n", t.Label, t.Value)
			} else {
				fmt.Printf("thermal: %-20s %.1f%s\n", t.Label, t.Value, t.Unit)
			}
		}
	}
	fmt.Println("fingerprint:", s.Fingerprint)
	fmt.Println("LTE:", s.LTE)
	fmt.Println("sleep:", s.Sleep)
}

func printMenu(r command.Runner, s Status) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println()
	fmt.Println("Actions:")
	fmt.Println("  1  Power saver")
	fmt.Println("  2  Balanced")
	fmt.Println("  3  Performance")
	fmt.Println("  4  Lock")
	fmt.Println("  5  Suspend")
	fmt.Println("  q  Close")
	fmt.Print("\nSelect: ")
	choice, _ := reader.ReadString('\n')
	switch strings.TrimSpace(strings.ToLower(choice)) {
	case "1":
		setProfile(r, []string{"power-saver"}, s.Profiles)
	case "2":
		setProfile(r, []string{"balanced"}, s.Profiles)
	case "3":
		setProfile(r, []string{"performance"}, s.Profiles)
	case "4":
		runCommand(r, "loginctl", "lock-session")
	case "5":
		runConfirmed(r, "Suspend", "systemctl", "suspend")
	case "q", "":
		fmt.Println("Closed.")
	default:
		fmt.Println("Unknown selection.")
	}
}

func batteries() []Battery {
	matches, _ := filepath.Glob("/sys/class/power_supply/BAT*")
	sort.Strings(matches)
	var out []Battery
	for _, p := range matches {
		name := filepath.Base(p)
		capacity := readInt(filepath.Join(p, "capacity"))
		status := strings.TrimSpace(readFile(filepath.Join(p, "status")))
		out = append(out, Battery{Name: name, Capacity: capacity, Status: status})
	}
	return out
}

func currentProfile(r command.Runner) string {
	if !r.Exists("powerprofilesctl") {
		return ""
	}
	out, err := r.Run("powerprofilesctl", "get")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func profiles(r command.Runner) []string {
	defaults := []string{"power-saver", "balanced", "performance"}
	if !r.Exists("powerprofilesctl") {
		return nil
	}
	out, err := r.Run("powerprofilesctl", "list")
	if err != nil {
		return defaults
	}
	var ps []string
	for _, p := range defaults {
		if strings.Contains(out, p) {
			ps = append(ps, p)
		}
	}
	if len(ps) == 0 {
		return defaults
	}
	return ps
}

func setProfile(r command.Runner, args []string, available []string) {
	if len(args) == 0 {
		fmt.Println("Usage: hyprglass laptop profile <power-saver|balanced|performance>")
		return
	}
	profile := args[0]
	if !slices.Contains(available, profile) {
		fmt.Println("Power profile is unavailable on this machine:", profile)
		return
	}
	runCommand(r, "powerprofilesctl", "set", profile)
}

func thermals(r command.Runner) []ThermalReading {
	var readings []ThermalReading
	if r.Exists("sensors") {
		out, err := r.Run("sensors", "-A")
		if err == nil {
			readings = append(readings, parseSensors(out)...)
		}
	}
	readings = append(readings, sysThermals()...)
	return compactReadings(readings)
}

var tempRE = regexp.MustCompile(`(?i)^([^:]+):\s*\+?([0-9]+(?:\.[0-9]+)?)\s*°?C`)
var fanRE = regexp.MustCompile(`(?i)^([^:]+):\s*([0-9]+(?:\.[0-9]+)?)\s*RPM`)

func parseSensors(out string) []ThermalReading {
	var readings []ThermalReading
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if m := tempRE.FindStringSubmatch(line); len(m) == 3 {
			v, _ := strconv.ParseFloat(m[2], 64)
			readings = append(readings, ThermalReading{Label: strings.TrimSpace(m[1]), Value: v, Unit: "C"})
		}
		if m := fanRE.FindStringSubmatch(line); len(m) == 3 {
			v, _ := strconv.ParseFloat(m[2], 64)
			readings = append(readings, ThermalReading{Label: strings.TrimSpace(m[1]), Value: v, Unit: "RPM"})
		}
	}
	return readings
}

func sysThermals() []ThermalReading {
	matches, _ := filepath.Glob("/sys/class/thermal/thermal_zone*/temp")
	sort.Strings(matches)
	var readings []ThermalReading
	for _, p := range matches {
		raw := readInt(p)
		if raw <= 0 {
			continue
		}
		label := filepath.Base(filepath.Dir(p))
		if t := strings.TrimSpace(readFile(filepath.Join(filepath.Dir(p), "type"))); t != "" {
			label = t
		}
		value := float64(raw)
		if raw > 1000 {
			value = value / 1000
		}
		readings = append(readings, ThermalReading{Label: label, Value: value, Unit: "C"})
	}
	return readings
}

func compactReadings(in []ThermalReading) []ThermalReading {
	seen := map[string]bool{}
	var out []ThermalReading
	for _, r := range in {
		key := r.Label + r.Unit
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, r)
		if len(out) >= 8 {
			break
		}
	}
	return out
}

func fingerprintStatus(r command.Runner) string {
	if !r.Exists("fprintd-enroll") || !r.Exists("fprintd-verify") {
		return "fprintd missing"
	}
	if r.Exists("fprintd-list") {
		user := os.Getenv("USER")
		args := []string{}
		if user != "" {
			args = append(args, user)
		}
		out, err := r.Run("fprintd-list", args...)
		if err == nil && strings.TrimSpace(out) != "" {
			return "available, enrollment data found"
		}
	}
	return "available"
}

func lteStatus(r command.Runner) string {
	if !r.Exists("mmcli") {
		return "ModemManager missing"
	}
	out, err := r.Run("mmcli", "-L")
	if err != nil {
		return "ModemManager unavailable"
	}
	if strings.Contains(out, "/Modem/") {
		return "modem detected"
	}
	return "no modem detected"
}

func sleepStatus(r command.Runner) string {
	if !r.Exists("systemctl") {
		return "systemctl missing"
	}
	out, err := r.Run("systemctl", "status", "hypridle", "--user", "--no-pager")
	if err == nil && strings.Contains(out, "active (running)") {
		return "hypridle active"
	}
	if r.Exists("loginctl") {
		out, err = r.Run("loginctl", "show-session", "self", "-p", "IdleHint")
		if err == nil && strings.TrimSpace(out) != "" {
			return "login session available"
		}
	}
	return "unknown"
}

func runConfirmed(r command.Runner, label, name string, args ...string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Type yes to confirm %s: ", label)
	confirm, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(confirm)) != "yes" {
		fmt.Println("Canceled.")
		return
	}
	runCommand(r, name, args...)
}

func runCommand(r command.Runner, name string, args ...string) {
	if !r.Exists(name) {
		fmt.Println(name, "is missing")
		return
	}
	out, err := r.Run(name, args...)
	if err != nil {
		fmt.Println("Command failed:", err)
	}
	fmt.Print(out)
}

func readFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

func readInt(path string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(readFile(path)))
	return n
}
