package display

import (
	"encoding/json"
	"fmt"
	"hyprglass/internal/command"
	"hyprglass/internal/tui"
)

type Monitor struct {
	Name        string  `json:"name"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	RefreshRate float64 `json:"refreshRate"`
	Scale       float64 `json:"scale"`
}

func ParseMonitors(out string) ([]Monitor, error) {
	var m []Monitor
	err := json.Unmarshal([]byte(out), &m)
	return m, err
}
func RunTUI(r command.Runner) {
	tui.Header("Display")
	if !r.Exists("hyprctl") {
		fmt.Println("hyprctl is missing or Hyprland is not installed.")
		return
	}
	out, err := r.Run("hyprctl", "monitors", "-j")
	if err != nil {
		fmt.Println("Not inside a usable Hyprland session:", err)
		return
	}
	ms, err := ParseMonitors(out)
	if err != nil {
		fmt.Println("could not parse hyprctl monitor JSON:", err)
		return
	}
	for _, m := range ms {
		fmt.Printf("%s  %dx%d @ %.2fHz scale %.2f\n", m.Name, m.Width, m.Height, m.RefreshRate, m.Scale)
	}
	fmt.Println("Monitor state shown. Use Hyprglass Settings to write the default scaling rule.")
}
