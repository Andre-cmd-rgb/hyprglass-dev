package command

import (
	"testing"
	"time"
)

func TestMockRunner(t *testing.T) {
	m := &MockRunner{Outputs: map[string]string{"echo ok": "ok"}, Present: map[string]bool{"echo": true}}
	out, err := m.Run("echo", "ok")
	if err != nil || out != "ok" || !m.Exists("echo") || m.Exists("missing") {
		t.Fatalf("bad mock")
	}
}

func TestTimeoutForSlowCommands(t *testing.T) {
	if got := timeoutFor("nmcli", "device", "wifi", "list"); got != 25*time.Second {
		t.Fatalf("nmcli wifi list timeout = %s", got)
	}
	if got := timeoutFor("mmcli", "-L"); got != 20*time.Second {
		t.Fatalf("mmcli timeout = %s", got)
	}
	if got := timeoutFor("python3", "scripts/generate-wallpaper.py"); got != 60*time.Second {
		t.Fatalf("python3 timeout = %s", got)
	}
	if got := timeoutFor("fprintd-enroll"); got != 90*time.Second {
		t.Fatalf("fprintd enroll timeout = %s", got)
	}
	if got := timeoutFor("true"); got != 15*time.Second {
		t.Fatalf("default timeout = %s", got)
	}
}
