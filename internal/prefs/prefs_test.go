package prefs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyDoesNotTouchDisplayConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	displayPath := filepath.Join(home, ".config", "hypr", "conf.d", "monitors.conf")
	manual := "# user tuned laptop panel\nmonitor = eDP-1, 3840x2400@60, 0x0, 1.75\n"
	if err := os.MkdirAll(filepath.Dir(displayPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(displayPath, []byte(manual), 0o644); err != nil {
		t.Fatal(err)
	}
	p := Default()
	p.Accent = "blue"
	if err := Apply(p); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(displayPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != manual {
		t.Fatalf("display config changed by safe Apply():\n%s", got)
	}
}

func TestApplyDisplayUpdatesManagedBlockOnly(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	displayPath := filepath.Join(home, ".config", "hypr", "conf.d", "monitors.conf")
	original := `# manual external monitor kept
monitor = HDMI-A-1, 1920x1080@60, 3840x0, 1

# >>> hyprglass managed display >>>
monitor = , preferred, auto, auto
# <<< hyprglass managed display <<<

# manual note kept
`
	if err := os.MkdirAll(filepath.Dir(displayPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(displayPath, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	p := Default()
	p.MonitorScale = "1.75"
	if err := ApplyDisplay(p); err != nil {
		t.Fatal(err)
	}
	gotBytes, err := os.ReadFile(displayPath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(gotBytes)
	if !strings.Contains(got, "monitor = HDMI-A-1, 1920x1080@60, 3840x0, 1") {
		t.Fatalf("manual monitor rule was not preserved:\n%s", got)
	}
	if !strings.Contains(got, "monitor = , preferred, auto, 1.75") {
		t.Fatalf("managed scale was not updated:\n%s", got)
	}
	if !strings.Contains(got, "# manual note kept") {
		t.Fatalf("manual trailing content was not preserved:\n%s", got)
	}
}

func TestApplyDisplayChangesOnlyScaleInsideManagedCustomRule(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	displayPath := filepath.Join(home, ".config", "hypr", "conf.d", "monitors.conf")
	original := `# >>> hyprglass managed display >>>
# User changed resolution and placement through Hyprland docs; Settings must keep them.
monitor = eDP-1, 3840x2400@60, 0x0, 1.75
# <<< hyprglass managed display <<<
`
	if err := os.MkdirAll(filepath.Dir(displayPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(displayPath, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	p := Default()
	p.MonitorScale = "2"
	if err := ApplyDisplay(p); err != nil {
		t.Fatal(err)
	}
	gotBytes, err := os.ReadFile(displayPath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(gotBytes)
	if !strings.Contains(got, "monitor = eDP-1, 3840x2400@60, 0x0, 2") {
		t.Fatalf("display scale update did not preserve custom resolution and position:\n%s", got)
	}
	if strings.Contains(got, "monitor = , preferred, auto") {
		t.Fatalf("display scale update replaced custom monitor rule with generic rule:\n%s", got)
	}
}

func TestApplyDisplayRefusesUnmarkedManualConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	displayPath := filepath.Join(home, ".config", "hypr", "conf.d", "monitors.conf")
	manual := "monitor = eDP-1, 3840x2400@60, 0x0, 1.75\n"
	if err := os.MkdirAll(filepath.Dir(displayPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(displayPath, []byte(manual), 0o644); err != nil {
		t.Fatal(err)
	}
	p := Default()
	p.MonitorScale = "2"
	if err := ApplyDisplay(p); err == nil {
		t.Fatal("expected manual display config to be protected")
	}
	got, err := os.ReadFile(displayPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != manual {
		t.Fatalf("manual display config changed despite refusal:\n%s", got)
	}
}

func TestApplyDisplayConvertsLegacyHyprglassRule(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	displayPath := filepath.Join(home, ".config", "hypr", "conf.d", "monitors.conf")
	legacy := "# Universal laptop-safe rule. Use Hyprglass Settings to generate a fixed scale.\nmonitor = , preferred, auto, auto\n"
	if err := os.MkdirAll(filepath.Dir(displayPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(displayPath, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	p := Default()
	p.MonitorScale = "1.5"
	if err := ApplyDisplay(p); err != nil {
		t.Fatal(err)
	}
	gotBytes, err := os.ReadFile(displayPath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(gotBytes)
	if !strings.Contains(got, displayStartMarker) || !strings.Contains(got, "monitor = , preferred, auto, 1.5") {
		t.Fatalf("legacy generated rule was not converted:\n%s", got)
	}
}
