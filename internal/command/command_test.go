package command

import "testing"

func TestMockRunner(t *testing.T) {
	m := &MockRunner{Outputs: map[string]string{"echo ok": "ok"}, Present: map[string]bool{"echo": true}}
	out, err := m.Run("echo", "ok")
	if err != nil || out != "ok" || !m.Exists("echo") || m.Exists("missing") {
		t.Fatalf("bad mock")
	}
}
