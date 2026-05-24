package wifi

import "testing"

func TestParseList(t *testing.T) {
	ns := ParseList("Home:87:WPA2\nCafe:40:\n")
	if len(ns) != 2 || ns[0].SSID != "Home" || ns[0].Signal != 87 || ns[0].Security != "WPA2" {
		t.Fatalf("bad parse: %#v", ns)
	}
}
