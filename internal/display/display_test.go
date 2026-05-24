package display

import "testing"

func TestParseMonitors(t *testing.T) {
	ms, err := ParseMonitors(`[{"name":"eDP-1","width":3840,"height":2400,"refreshRate":60,"scale":1.75}]`)
	if err != nil || len(ms) != 1 || ms[0].Scale != 1.75 {
		t.Fatalf("bad parse %#v %v", ms, err)
	}
}
