package wifi

import "testing"

func TestParseList(t *testing.T) {
	ns := ParseList("Home|87|WPA2\nCafe|40|\n")
	if len(ns) != 2 || ns[0].SSID != "Home" || ns[0].Signal != 87 || ns[0].Security != "WPA2" {
		t.Fatalf("bad parse: %#v", ns)
	}
	// SSID containing a colon must parse correctly with | separator
	ns2 := ParseList("My:Net|72|WPA3\n")
	if len(ns2) != 1 || ns2[0].SSID != "My:Net" {
		t.Fatalf("colon-in-ssid parse failed: %#v", ns2)
	}
}
