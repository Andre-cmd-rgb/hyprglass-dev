package prefs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Preferences struct {
	ThemeMode       string `json:"themeMode"`
	Accent          string `json:"accent"`
	KeyboardLayout  string `json:"keyboardLayout"`
	KeyboardVariant string `json:"keyboardVariant"`
	MonitorScale    string `json:"monitorScale"`
	ModemAPN        string `json:"modemApn"`
	ModemPINSet     bool   `json:"modemPinSet"`
}

type Palette struct {
	Name    string
	Accent  string
	Accent2 string
	BG      string
	Panel   string
	Text    string
	Muted   string
	Border  string
	Warn    string
	Danger  string
}

func Default() Preferences {
	return Preferences{
		ThemeMode:      "dark",
		Accent:         "graphite",
		KeyboardLayout: "us",
		MonitorScale:   "auto",
	}
}

func Path() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".hyprglass-preferences.json"
	}
	return filepath.Join(home, ".config", "hyprglass", "preferences.json")
}

func Load() Preferences {
	p := Default()
	b, err := os.ReadFile(Path())
	if err != nil {
		return p
	}
	_ = json.Unmarshal(b, &p)
	p.Normalize()
	return p
}

func Save(p Preferences) error {
	p.Normalize()
	path := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

func (p *Preferences) Normalize() {
	p.ThemeMode = strings.ToLower(strings.TrimSpace(p.ThemeMode))
	if p.ThemeMode != "light" {
		p.ThemeMode = "dark"
	}
	p.Accent = strings.ToLower(strings.TrimSpace(p.Accent))
	if _, ok := accentHex[p.Accent]; !ok {
		p.Accent = "graphite"
	}
	p.KeyboardLayout = strings.TrimSpace(p.KeyboardLayout)
	if p.KeyboardLayout == "" {
		p.KeyboardLayout = "us"
	}
	p.KeyboardVariant = strings.TrimSpace(p.KeyboardVariant)
	p.MonitorScale = strings.ToLower(strings.TrimSpace(p.MonitorScale))
	if p.MonitorScale == "" {
		p.MonitorScale = "auto"
	}
	if !validScale(p.MonitorScale) {
		p.MonitorScale = "auto"
	}
	p.ModemAPN = strings.TrimSpace(p.ModemAPN)
}

func validScale(v string) bool {
	switch v {
	case "auto", "1", "1.10", "1.15", "1.20", "1.25", "1.33", "1.50", "1.5", "1.60", "1.67", "1.75", "2", "2.0":
		return true
	default:
		return false
	}
}

var accentHex = map[string][2]string{
	"graphite": {"8ea8ff", "b8c7ff"},
	"blue":     {"5aa7ff", "98c8ff"},
	"cyan":     {"4cc9f0", "a7ecff"},
	"green":    {"7bd88f", "b8f2c2"},
	"orange":   {"ffb86b", "ffd0a3"},
	"red":      {"ff7b87", "ffb1b9"},
	"pink":     {"ff8bd1", "ffc0e7"},
	"purple":   {"b69cff", "d5c6ff"},
}

func AccentNames() []string {
	return []string{"graphite", "blue", "cyan", "green", "orange", "red", "pink", "purple"}
}

func PaletteFor(p Preferences) Palette {
	p.Normalize()
	pair := accentHex[p.Accent]
	if p.ThemeMode == "light" {
		return Palette{
			Name: p.Accent, Accent: pair[0], Accent2: pair[1],
			BG: "f7f7f2", Panel: "fffffff0", Text: "16181d", Muted: "68707d", Border: "00000018",
			Warn: "b7791f", Danger: "d93b4a",
		}
	}
	return Palette{
		Name: p.Accent, Accent: pair[0], Accent2: pair[1],
		BG: "0f1115", Panel: "1b1f2acc", Text: "f2f2ec", Muted: "a5adba", Border: "ffffff24",
		Warn: "ffd166", Danger: "ff7b87",
	}
}

func Apply(p Preferences) error {
	p.Normalize()
	pal := PaletteFor(p)
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	writes := map[string]string{
		filepath.Join(home, ".config", "hypr", "conf.d", "theme.conf"):    hyprTheme(pal),
		filepath.Join(home, ".config", "hypr", "conf.d", "input.conf"):    inputConfig(p),
		filepath.Join(home, ".config", "hypr", "conf.d", "monitors.conf"): monitorConfig(p),
		filepath.Join(home, ".config", "waybar", "style.css"):             waybarCSS(pal, p.ThemeMode),
		filepath.Join(home, ".config", "mako", "config"):                  makoConfig(pal),
		filepath.Join(home, ".config", "fuzzel", "fuzzel.ini"):            fuzzelConfig(pal),
		filepath.Join(home, ".config", "gtk-3.0", "settings.ini"):         gtkSettings(p),
		filepath.Join(home, ".config", "gtk-4.0", "settings.ini"):         gtkSettings(p),
	}
	for path, data := range writes {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func hyprTheme(p Palette) string {
	return fmt.Sprintf(`# Hyprglass generated theme. Edit through hyprglass settings.
$hg_bg = rgba(%see)
$hg_bg_soft = rgba(%sdd)
$hg_panel = rgba(%s)
$hg_panel_light = rgba(fffffff0)
$hg_text = rgba(%see)
$hg_text_muted = rgba(%see)
$hg_accent = rgba(%see)
$hg_accent_2 = rgba(%see)
$hg_border = rgba(%s)
`, p.BG, p.BG, p.Panel, p.Text, p.Muted, p.Accent, p.Accent2, strings.TrimPrefix(p.Border, "#"))
}

func inputConfig(p Preferences) string {
	return fmt.Sprintf(`input {
    kb_layout  = %s
    kb_variant = %s
    kb_options =
    follow_mouse = 1
    sensitivity  = 0

    touchpad {
        natural_scroll       = true
        disable_while_typing = true
        tap-to-click         = true
        scroll_factor        = 1.0
        drag_lock            = false
    }
}

gestures {
    gesture = 3, horizontal, workspace
}
`, p.KeyboardLayout, p.KeyboardVariant)
}

func monitorConfig(p Preferences) string {
	scale := p.MonitorScale
	if scale == "" {
		scale = "auto"
	}
	return fmt.Sprintf(`# Hyprglass generated display rule. Edit through hyprglass settings.
# Uses preferred panel resolution and automatic monitor placement.
monitor = , preferred, auto, %s
`, scale)
}

func waybarCSS(p Palette, mode string) string {
	return fmt.Sprintf(`* {
  border: none;
  border-radius: 0;
  font-family: "JetBrainsMono Nerd Font", "JetBrainsMono Nerd Font Mono", "Symbols Nerd Font", monospace;
  font-size: 12px;
  min-height: 0;
}

window#waybar {
  background: #%s;
  border: 1px solid #%s;
  border-radius: 13px;
  color: #%s;
}

.modules-left,
.modules-center,
.modules-right {
  background: transparent;
  padding: 0 10px;
}

#custom-logo {
  color: #%s;
  font-weight: 700;
  padding: 0 10px 0 4px;
}

#workspaces { padding: 0 4px; }

#workspaces button {
  background: transparent;
  color: #%s;
  padding: 0 9px;
  margin: 0 1px;
}

#workspaces button.active {
  color: #%s;
  background: #%s;
  border-radius: 9px;
}

#clock,
#network,
#custom-bluetooth,
#pulseaudio,
#backlight,
#battery,
#custom-settings,
#custom-power {
  color: #%s;
  padding: 0 9px;
  margin: 0 1px;
}

#clock { color: #%s; }
#custom-settings { color: #%s; }
#custom-power { color: #%s; }
#network.disconnected,
#battery.warning { color: #%s; }
#battery.critical { color: #%s; }
#custom-bluetooth.off { color: #%s; }
#custom-bluetooth.on { color: #%s; }
`, panelCSS(p.Panel), p.Border, p.Text, p.Text, p.Muted, p.BG, p.Accent, p.Text, p.Muted, p.Accent, p.Danger, p.Warn, p.Danger, p.Muted, p.Accent)
}

func panelCSS(v string) string {
	if len(v) == 8 {
		return v
	}
	return v + "d8"
}

func makoConfig(p Palette) string {
	return fmt.Sprintf(`font=JetBrainsMono Nerd Font 11
background-color=#%s
text-color=#%s
border-color=#%s
progress-color=over #%s
border-size=1
border-radius=14
padding=12
margin=12
width=360
height=110
default-timeout=5000
anchor=top-right
icons=1
max-icon-size=32
layer=overlay
`, panelCSS(p.Panel), p.Text, p.Border, p.Accent)
}

func fuzzelConfig(p Palette) string {
	return fmt.Sprintf(`[main]
font=JetBrainsMono Nerd Font:size=13
prompt=""
width=44
lines=10
horizontal-pad=18
vertical-pad=14
inner-pad=8
terminal=kitty
layer=overlay

[colors]
background=%s
text=%sff
match=%sff
selection=%see
selection-text=%sff
selection-match=%sff
border=%s

[border]
radius=16
width=1
`, panelCSS(p.Panel), p.Text, p.Accent, p.BG, p.Text, p.Accent2, p.Border)
}

func gtkSettings(p Preferences) string {
	prefer := "prefer-dark"
	theme := "Adwaita-dark"
	if p.ThemeMode == "light" {
		prefer = "prefer-light"
		theme = "Adwaita"
	}
	return fmt.Sprintf(`[Settings]
gtk-theme-name=%s
gtk-application-prefer-dark-theme=%d
gtk-font-name=Inter 10
gtk-icon-theme-name=Adwaita
# color-scheme: %s
`, theme, boolInt(p.ThemeMode == "dark"), prefer)
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
