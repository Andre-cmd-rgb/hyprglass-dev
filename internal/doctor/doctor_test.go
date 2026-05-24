package doctor

import (
	"encoding/json"
	"hyprglass/internal/command"
	"testing"
)

func TestDoctorJSON(t *testing.T) {
	r := Run(&command.MockRunner{Present: map[string]bool{"go": true}})
	b, err := json.Marshal(r)
	if err != nil || !json.Valid(b) || len(r.Checks) == 0 {
		t.Fatalf("invalid doctor")
	}
}
