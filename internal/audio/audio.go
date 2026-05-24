package audio

import (
	"fmt"
	"hyprglass/internal/command"
	"hyprglass/internal/tui"
	"strings"
)

func ParseStatus(out string) []string {
	var xs []string
	for _, l := range strings.Split(out, "\n") {
		l = strings.TrimSpace(l)
		if strings.Contains(l, "Sink") || strings.Contains(l, "Source") || strings.Contains(l, "Default") {
			xs = append(xs, l)
		}
	}
	return xs
}
func RunTUI(r command.Runner) {
	tui.Header("Audio")
	if !r.Exists("wpctl") {
		fmt.Println("wpctl is missing. Install wireplumber.")
		return
	}
	out, err := r.Run("wpctl", "status")
	if err != nil {
		fmt.Println("PipeWire/WirePlumber status unavailable:", err)
		return
	}
	for _, l := range ParseStatus(out) {
		fmt.Println(l)
	}
	fmt.Println("Use wpctl set-volume, set-mute, and set-default for direct actions.")
}
