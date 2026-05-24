package command

import (
	"context"
	"errors"
	"os/exec"
	"time"
)

type Runner interface {
	Run(name string, args ...string) (string, error)
	Exists(name string) bool
}
type RealRunner struct{}

func (RealRunner) Run(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutFor(name, args...))
	defer cancel()
	b, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return string(b), ctx.Err()
	}
	return string(b), err
}
func (RealRunner) Exists(name string) bool { _, err := exec.LookPath(name); return err == nil }

func timeoutFor(name string, args ...string) time.Duration {
	switch name {
	case "nmcli":
		if hasArgs(args, "device", "wifi", "list") || hasArgs(args, "device", "wifi", "rescan") {
			return 25 * time.Second
		}
		return 15 * time.Second
	case "mmcli":
		return 20 * time.Second
	case "hyprctl", "bluetoothctl", "wpctl", "systemctl", "loginctl":
		return 15 * time.Second
	case "python3":
		return 60 * time.Second
	default:
		return 15 * time.Second
	}
}

func hasArgs(args []string, want ...string) bool {
	if len(want) == 0 || len(args) < len(want) {
		return false
	}
	for i := 0; i <= len(args)-len(want); i++ {
		ok := true
		for j, w := range want {
			if args[i+j] != w {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

type MockRunner struct {
	Outputs map[string]string
	Errs    map[string]error
	Present map[string]bool
	Calls   []string
}

func (m *MockRunner) Run(name string, args ...string) (string, error) {
	key := name
	for _, a := range args {
		key += " " + a
	}
	m.Calls = append(m.Calls, key)
	if e, ok := m.Errs[key]; ok {
		return m.Outputs[key], e
	}
	return m.Outputs[key], nil
}
func (m *MockRunner) Exists(name string) bool {
	if m.Present == nil {
		return true
	}
	return m.Present[name]
}
