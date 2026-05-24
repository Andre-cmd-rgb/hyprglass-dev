package tui

import (
	"bufio"
	"fmt"
	"hyprglass/internal/doctor"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func Header(title string) { fmt.Printf("\033[38;5;147m╭─ Hyprglass :: %s ─╮\033[0m\n", title) }

func PrintChecks(r doctor.Result) {
	Header("Doctor")
	fmt.Println("status:", r.Status)
	for _, c := range r.Checks {
		fmt.Printf("%-34s %-5s %s\n", c.Name, c.Status, c.Message)
	}
}

func Clear() {
	if os.Getenv("TERM") != "" {
		fmt.Print("\033[H\033[2J")
	}
}

func ReadLine(r *bufio.Reader) string {
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

func Pause(r *bufio.Reader) {
	fmt.Print("\nPress Enter to continue.")
	_, _ = r.ReadString('\n')
}

// Launch starts name as a detached background process, ignoring errors.
func Launch(name string) {
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
