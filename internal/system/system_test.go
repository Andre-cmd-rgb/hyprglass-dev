package system

import (
	"hyprglass/internal/command"
	"testing"
)

func TestCollectUsesRunnerPresence(t *testing.T) {
	r := &command.MockRunner{
		Present: map[string]bool{"uname": true, "pacman": true, "chwd": true, "cachyos-rate-mirrors": true},
		Outputs: map[string]string{
			"uname -r":                "6.17.0-cachyos\n",
			"pacman -Q linux-cachyos": "linux-cachyos 6.17.0\n",
		},
	}
	st := Collect(r)
	if st.Kernel != "6.17.0-cachyos" {
		t.Fatalf("kernel=%q", st.Kernel)
	}
	if !st.Pacman || !st.CHWD || !st.CachyRateMirrors {
		t.Fatalf("expected tools present: %+v", st)
	}
}
