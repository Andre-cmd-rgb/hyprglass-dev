package bluetooth

import "testing"

func TestParseDevices(t *testing.T) {
	ds := ParseDevices("Device AA:BB Keyboard Pro\n")
	if len(ds) != 1 || ds[0].Name != "Keyboard Pro" {
		t.Fatalf("bad parse %#v", ds)
	}
}
