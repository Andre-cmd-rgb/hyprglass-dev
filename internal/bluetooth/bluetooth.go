package bluetooth

import (
	"encoding/json"
	"fmt"
	"hyprglass/internal/command"
	"hyprglass/internal/tui"
	"strings"
)

type Device struct{ MAC, Name string }

func ParseDevices(out string) []Device {
	var ds []Device
	for _, l := range strings.Split(out, "\n") {
		f := strings.Fields(l)
		if len(f) >= 3 && (f[0] == "Device" || f[0] == "Controller") {
			ds = append(ds, Device{MAC: f[1], Name: strings.Join(f[2:], " ")})
		}
	}
	return ds
}
func PrintWaybarStatus(r command.Runner) {
	txt := "󰂯"
	cls := "off"
	if r.Exists("bluetoothctl") {
		out, _ := r.Run("bluetoothctl", "show")
		if strings.Contains(out, "Powered: yes") {
			cls = "on"
		}
	}
	b, _ := json.Marshal(map[string]string{"text": txt, "class": cls, "tooltip": "Bluetooth " + cls})
	fmt.Println(string(b))
}
func RunTUI(r command.Runner) {
	tui.Header("Bluetooth")
	if !r.Exists("bluetoothctl") {
		fmt.Println("bluetoothctl is missing. Install bluez-utils.")
		return
	}
	out, err := r.Run("bluetoothctl", "show")
	if err != nil {
		fmt.Println("bluetooth adapter unavailable or service stopped:", err)
		return
	}
	fmt.Println(strings.TrimSpace(out))
	devs, _ := r.Run("bluetoothctl", "devices")
	fmt.Println("\ndevices:")
	for _, d := range ParseDevices(devs) {
		fmt.Printf("%s  %s\n", d.MAC, d.Name)
	}
	fmt.Println("Pairing may require an agent. Use bluetoothctl interactively if this screen cannot complete it.")
}
