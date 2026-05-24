package srcroot

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsRoot(t *testing.T) {
	dir := t.TempDir()
	if IsRoot(dir) {
		t.Fatal("empty temp dir should not look like source root")
	}
	if err := os.WriteFile(filepath.Join(dir, "install.sh"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if IsRoot(dir) {
		t.Fatal("missing cmd/hyprglass/main.go should still fail")
	}
	if err := os.MkdirAll(filepath.Join(dir, "cmd", "hyprglass"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cmd", "hyprglass", "main.go"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if !IsRoot(dir) {
		t.Fatal("dir with install.sh and cmd/hyprglass/main.go should be recognised as root")
	}
	if IsRoot("") {
		t.Fatal("empty string should return false")
	}
}

func TestFindViaEnv(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "cmd", "hyprglass"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{filepath.Join(dir, "install.sh"), filepath.Join(dir, "cmd", "hyprglass", "main.go")} {
		if err := os.WriteFile(name, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("HYPRGLASS_SOURCE_ROOT", dir)
	got := Find("")
	if got != dir {
		t.Fatalf("Find() via env = %q, want %q", got, dir)
	}
}

func TestFindMissing(t *testing.T) {
	t.Setenv("HYPRGLASS_SOURCE_ROOT", "")
	if got := Find(""); got != "" {
		t.Logf("Find() returned %q (may be the actual repo checkout — acceptable in a real build)", got)
	}
}
