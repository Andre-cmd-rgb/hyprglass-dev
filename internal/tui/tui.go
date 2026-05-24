package tui

import (
	"fmt"
	"hyprglass/internal/doctor"
)

func Header(title string) { fmt.Printf("\033[38;5;147m╭─ Hyprglass :: %s ─╮\033[0m\n", title) }
func PrintChecks(r doctor.Result) {
	Header("Doctor")
	fmt.Println("status:", r.Status)
	for _, c := range r.Checks {
		fmt.Printf("%-34s %-5s %s\n", c.Name, c.Status, c.Message)
	}
}
