package audio

import "testing"

func TestParseStatus(t *testing.T) {
	xs := ParseStatus("Audio\n ├─ Sinks:\n │  * 42. Speaker\n └─ Sources:\n")
	if len(xs) < 2 {
		t.Fatalf("bad parse %#v", xs)
	}
}
