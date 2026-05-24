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
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	b, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return string(b), ctx.Err()
	}
	return string(b), err
}
func (RealRunner) Exists(name string) bool { _, err := exec.LookPath(name); return err == nil }

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
