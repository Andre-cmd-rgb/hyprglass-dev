package prefs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

type ApplyOptions struct {
	Visuals bool
	Input   bool
	Display bool
}

const (
	displayStartMarker = "# >>> hyprglass managed display >>>"
	displayEndMarker   = "# <<< hyprglass managed display <<<"
)

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
	"graphite": {"9aa0aa", "d8dce2"},
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
			BG: "f7f7f2", Panel: "fffffff2", Text: "16181d", Muted: "68707d", Border: "0000001f",
			Warn: "b7791f", Danger: "d93b4a",
		}
	}
	return Palette{
		Name: p.Accent, Accent: pair[0], Accent2: pair[1],
		BG: "0b0d10", Panel: "161922d1", Text: "f2f2ec", Muted: "a5adba", Border: "ffffff24",
		Warn: "ffd166", Danger: "ff7b87",
	}
}

// Apply is intentionally display-safe. It refreshes theme, Waybar, launcher,
// GTK, Mako, and input settings, but it does not touch monitor rules. Display
// changes are applied only by ApplyDisplay or ApplyAll so a quick appearance
// change cannot destroy a custom laptop/external-monitor layout.
func Apply(p Preferences) error {
	return ApplyWithOptions(p, ApplyOptions{Visuals: true, Input: true})
}

func ApplyVisuals(p Preferences) error {
	return ApplyWithOptions(p, ApplyOptions{Visuals: true})
}

func ApplyInput(p Preferences) error {
	return ApplyWithOptions(p, ApplyOptions{Input: true})
}

func ApplyDisplay(p Preferences) error {
	return ApplyWithOptions(p, ApplyOptions{Display: true})
}

func ApplyDisplayAndInput(p Preferences) error {
	return ApplyWithOptions(p, ApplyOptions{Display: true, Input: true})
}

func ApplyAll(p Preferences) error {
	return ApplyWithOptions(p, ApplyOptions{Visuals: true, Input: true, Display: true})
}

func ApplyWithOptions(p Preferences, opts ApplyOptions) error {
	p.Normalize()
	pal := PaletteFor(p)
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	writes := map[string]string{}
	if opts.Visuals {
		writes[filepath.Join(home, ".config", "hypr", "conf.d", "theme.conf")] = hyprTheme(pal)
		writes[filepath.Join(home, ".config", "waybar", "style.css")] = waybarCSS(pal, p.ThemeMode)
		writes[filepath.Join(home, ".config", "mako", "config")] = makoConfig(pal)
		writes[filepath.Join(home, ".config", "fuzzel", "fuzzel.ini")] = fuzzelConfig(pal)
		writes[filepath.Join(home, ".config", "gtk-3.0", "settings.ini")] = gtkSettings(p)
		writes[filepath.Join(home, ".config", "gtk-4.0", "settings.ini")] = gtkSettings(p)
	}
	if opts.Input {
		writes[filepath.Join(home, ".config", "hypr", "conf.d", "input.conf")] = inputConfig(p)
	}
	for path, data := range writes {
		if err := writeFile(path, data); err != nil {
			return err
		}
	}
	if opts.Display {
		path := filepath.Join(home, ".config", "hypr", "conf.d", "monitors.conf")
		if err := writeManagedDisplayScale(path, p.MonitorScale); err != nil {
			return err
		}
	}
	return nil
}

func writeFile(path, data string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(data), 0o644)
}

func writeManagedDisplayScale(path, scale string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	scale = strings.TrimSpace(scale)
	if scale == "" || !validScale(scale) {
		scale = "auto"
	}
	defaultBlock := monitorConfigWithScale(scale)
	oldBytes, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return os.WriteFile(path, []byte(displayFile(defaultBlock)), 0o644)
	}
	if err != nil {
		return err
	}
	old := string(oldBytes)
	start := strings.Index(old, displayStartMarker)
	end := strings.Index(old, displayEndMarker)
	if start >= 0 && end >= 0 && end > start {
		end += len(displayEndMarker)
		block := old[start:end]
		updatedBlock := updateDisplayBlockScale(block, scale)
		updated := strings.TrimRight(old[:start], " \t\r\n") + "\n" + strings.TrimRight(updatedBlock, "\n") + "\n" + strings.TrimLeft(old[end:], " \t\r\n")
		return os.WriteFile(path, []byte(ensureTrailingNewline(updated)), 0o644)
	}
	if isLegacyHyprglassDisplay(old) || strings.TrimSpace(old) == "" {
		return os.WriteFile(path, []byte(displayFile(defaultBlock)), 0o644)
	}
	if hasActiveMonitorRule(old) {
		return fmt.Errorf("refusing to overwrite manual display config at %s; add a %q block around the line you want Hyprglass to scale, or edit the file manually", path, displayStartMarker)
	}
	updated := strings.TrimRight(old, " \t\r\n") + "\n\n" + strings.TrimRight(defaultBlock, "\n") + "\n"
	return os.WriteFile(path, []byte(updated), 0o644)
}

func updateDisplayBlockScale(block, scale string) string {
	lines := strings.Split(block, "\n")
	changed := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || !strings.HasPrefix(trimmed, "monitor") {
			continue
		}
		lhs, rhs, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(lhs) != "monitor" {
			continue
		}
		comment := ""
		if hash := strings.Index(rhs, "#"); hash >= 0 {
			comment = rhs[hash:]
			rhs = rhs[:hash]
		}
		parts := strings.Split(rhs, ",")
		for len(parts) < 4 {
			parts = append(parts, "")
		}
		parts[3] = replaceFieldKeepingIndent(parts[3], scale)
		lines[i] = lhs + "=" + strings.Join(parts, ",") + restoreInlineComment(comment)
		changed = true
	}
	if !changed {
		insert := strings.Split(monitorConfigWithScale(scale), "\n")
		// Keep the existing markers and place the default monitor line between them.
		var out []string
		inserted := false
		for _, line := range lines {
			out = append(out, line)
			if !inserted && strings.TrimSpace(line) == displayStartMarker {
				for _, il := range insert {
					if strings.TrimSpace(il) == "" || strings.TrimSpace(il) == displayStartMarker || strings.TrimSpace(il) == displayEndMarker || strings.HasPrefix(strings.TrimSpace(il), "#") {
						continue
					}
					out = append(out, il)
				}
				inserted = true
			}
		}
		lines = out
	}
	return strings.Join(lines, "\n")
}

func replaceFieldKeepingIndent(field, value string) string {
	leading := field[:len(field)-len(strings.TrimLeft(field, " \t"))]
	trailing := field[len(strings.TrimRight(field, " \t")):]
	if leading == "" {
		leading = " "
	}
	return leading + value + trailing
}

func restoreInlineComment(comment string) string {
	if comment == "" {
		return ""
	}
	if strings.HasPrefix(comment, " ") || strings.HasPrefix(comment, "\t") {
		return comment
	}
	return " " + comment
}

func displayFile(block string) string {
	return `# Hyprglass display configuration.
# This file is safe for custom laptop/external-monitor layouts.
# Hyprglass Settings only rewrites the marked block below.
# Put manual monitor rules outside the block, or remove the block and manage display yourself.

` + strings.TrimRight(block, "\n") + "\n"
}

func isLegacyHyprglassDisplay(data string) bool {
	if !strings.Contains(data, "Hyprglass generated display rule") && !strings.Contains(data, "Universal laptop-safe rule") {
		return false
	}
	active := activeMonitorRules(data)
	if len(active) == 0 {
		return true
	}
	for _, line := range active {
		normalized := strings.ReplaceAll(line, " ", "")
		if !strings.HasPrefix(normalized, "monitor=,preferred,auto,") {
			return false
		}
	}
	return true
}

func hasActiveMonitorRule(data string) bool {
	return len(activeMonitorRules(data)) > 0
}

func activeMonitorRules(data string) []string {
	var rules []string
	for _, line := range strings.Split(data, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "monitor") {
			lhs := strings.TrimSpace(strings.SplitN(trimmed, "=", 2)[0])
			if lhs == "monitor" || strings.HasPrefix(lhs, "monitorv2") {
				rules = append(rules, trimmed)
			}
		}
	}
	return rules
}

func ensureTrailingNewline(s string) string {
	if strings.HasSuffix(s, "\n") {
		return s
	}
	return s + "\n"
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

func monitorConfigWithScale(scale string) string {
	if scale == "" {
		scale = "auto"
	}
	return fmt.Sprintf(`%s
# Preferred panel resolution, automatic placement, configurable scale.
# Settings changes only the scale field on monitor lines inside this block.
monitor = , preferred, auto, %s
%s
`, displayStartMarker, scale, displayEndMarker)
}

func waybarCSS(p Palette, mode string) string {
	return fmt.Sprintf(`* {
  border: none;
  border-radius: 0;
  font-family: "JetBrainsMono Nerd Font", "JetBrainsMonoNL Nerd Font", "JetBrainsMono Nerd Font Propo", "JetBrainsMono Nerd Font Mono", "Symbols Nerd Font Mono", "Symbols Nerd Font", "Noto Sans", "DejaVu Sans", sans-serif;
  font-size: 12px;
  min-height: 0;
}

window#waybar {
  background: %s;
  border: 1px solid %s;
  border-radius: 13px;
  color: %s;
}

.modules-left,
.modules-center,
.modules-right {
  background: transparent;
  padding: 0 10px;
}

#custom-logo {
  color: %s;
  font-weight: 700;
  padding: 0 10px 0 4px;
}

#workspaces { padding: 0 4px; }

#workspaces button {
  background: transparent;
  color: %s;
  padding: 0 9px;
  margin: 0 1px;
}

#workspaces button.active {
  color: %s;
  background: %s;
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
  color: %s;
  padding: 0 9px;
  margin: 0 1px;
}

#clock { color: %s; }
#network,
#custom-bluetooth,
#pulseaudio,
#backlight,
#battery,
#custom-settings,
#custom-power {
  font-family: "JetBrainsMono Nerd Font", "JetBrainsMonoNL Nerd Font", "Symbols Nerd Font Mono", "Symbols Nerd Font", "Noto Sans", "DejaVu Sans", sans-serif;
}

#custom-bluetooth,
#custom-settings,
#custom-power {
  font-size: 13px;
}

#custom-settings { color: %s; }
#custom-power { color: %s; }
#network.disconnected,
#battery.warning { color: %s; }
#battery.critical { color: %s; }
#custom-bluetooth.off { color: %s; }
#custom-bluetooth.on { color: %s; }
`, cssColor(panelCSS(p.Panel)), cssColor(p.Border), cssColor(p.Text), cssColor(p.Text), cssColor(p.Muted), cssColor(p.BG), cssColor(p.Accent), cssColor(p.Text), cssColor(p.Muted), cssColor(p.Accent), cssColor(p.Danger), cssColor(p.Warn), cssColor(p.Danger), cssColor(p.Muted), cssColor(p.Accent))
}

func cssColor(v string) string {
	h := strings.TrimPrefix(strings.TrimSpace(v), "#")
	if len(h) == 6 {
		return "#" + h
	}
	if len(h) == 8 {
		r, er := strconv.ParseInt(h[0:2], 16, 64)
		g, eg := strconv.ParseInt(h[2:4], 16, 64)
		b, eb := strconv.ParseInt(h[4:6], 16, 64)
		a, ea := strconv.ParseInt(h[6:8], 16, 64)
		if er == nil && eg == nil && eb == nil && ea == nil {
			return fmt.Sprintf("rgba(%d, %d, %d, %.2f)", r, g, b, float64(a)/255.0)
		}
	}
	return "#" + h
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
