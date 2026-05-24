package lte

import (
	"fmt"
	"hyprglass/internal/command"
	"hyprglass/internal/tui"
	"regexp"
	"strings"
)

type Modem struct{ Path, Label string }

var re = regexp.MustCompile(`/org/freedesktop/ModemManager1/Modem/[0-9]+`)

func ParseList(out string) []Modem {
	var ms []Modem
	for _, l := range strings.Split(out, "\n") {
		p := re.FindString(l)
		if p != "" {
			ms = append(ms, Modem{Path: p, Label: strings.TrimSpace(l)})
		}
	}
	return ms
}
func RunTUI(r command.Runner) {
	tui.Header("LTE")
	if !r.Exists("mmcli") {
		fmt.Println("mmcli is missing. Install modemmanager.")
		return
	}
	out, err := r.Run("mmcli", "-L")
	if err != nil {
		fmt.Println("ModemManager unavailable:", err)
		return
	}
	ms := ParseList(out)
	if len(ms) == 0 {
		fmt.Println("No modem detected.")
		return
	}
	for _, m := range ms {
		fmt.Println(m.Label)
		idx := strings.TrimPrefix(m.Path, "/org/freedesktop/ModemManager1/Modem/")
		info, _ := r.Run("mmcli", "-m", idx)
		fmt.Println(info)
	}
	fmt.Println("Connect example: mmcli -m <index> --simple-connect=\"apn=<apn>\". Hyprglass will not restart networking without confirmation.")
}
