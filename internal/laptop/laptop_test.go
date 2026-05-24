package laptop

import "testing"

func TestParseSensors(t *testing.T) {
	readings := parseSensors(`coretemp-isa-0000
Package id 0:  +47.0°C  (high = +100.0°C, crit = +100.0°C)
fan1:        2400 RPM
`)
	if len(readings) != 2 {
		t.Fatalf("expected 2 readings, got %#v", readings)
	}
	if readings[0].Label != "Package id 0" || readings[0].Value != 47 || readings[0].Unit != "C" {
		t.Fatalf("bad temp parse: %#v", readings[0])
	}
	if readings[1].Label != "fan1" || readings[1].Value != 2400 || readings[1].Unit != "RPM" {
		t.Fatalf("bad fan parse: %#v", readings[1])
	}
}
