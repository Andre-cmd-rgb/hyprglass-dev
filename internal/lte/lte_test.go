package lte

import "testing"

func TestParseList(t *testing.T) {
	ms := ParseList("/org/freedesktop/ModemManager1/Modem/1 [foxconn] Qualcomm Snapdragon X55 5G")
	if len(ms) != 1 || ms[0].Path == "" {
		t.Fatalf("bad parse %#v", ms)
	}
}
