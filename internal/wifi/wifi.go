package wifi

import (
	"fmt"
	"hyprglass/internal/command"
	"hyprglass/internal/tui"
	"strconv"
	"strings"
)

type Network struct {
	SSID, Security string
	Signal         int
}

func ParseList(out string) []Network {
	var ns []Network
	for _, l := range strings.Split(out, "\n") {
		l = strings.TrimSpace(strings.TrimPrefix(l, "*"))
		if l == "" || strings.HasPrefix(l, "SSID") {
			continue
		}
		parts := strings.SplitN(l, "|", 3)
		if len(parts) < 2 {
			continue
		}
		sig, _ := strconv.Atoi(parts[1])
		sec := ""
		if len(parts) > 2 {
			sec = parts[2]
		}
		ns = append(ns, Network{SSID: parts[0], Signal: sig, Security: sec})
	}
	return ns
}
func RunTUI(r command.Runner) {
	tui.Header("Wi-Fi")
	if !r.Exists("nmcli") {
		fmt.Println("nmcli is missing. Install networkmanager.")
		return
	}
	radio, _ := r.Run("nmcli", "radio", "wifi")
	fmt.Println("radio:", strings.TrimSpace(radio))
	active, _ := r.Run("nmcli", "-t", "-f", "NAME,TYPE,DEVICE", "connection", "show", "--active")
	fmt.Println("active connections:\n" + strings.TrimSpace(active))
	out, err := r.Run("nmcli", "-t", "--escape", "no", "--separator", "|", "-f", "SSID,SIGNAL,SECURITY", "device", "wifi", "list")
	if err != nil {
		fmt.Println("could not list Wi-Fi networks:", err)
		return
	}
	for _, n := range ParseList(out) {
		lock := ""
		if n.Security != "" && n.Security != "--" {
			lock = " locked"
		}
		fmt.Printf("%-28s %3d%%%s\n", n.SSID, n.Signal, lock)
	}
	fmt.Println("Actions: use nmcli device wifi rescan/connect/disconnect. Password entry stays explicit for safety.")
}
