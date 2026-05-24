package srcroot

import (
	"os"
	"path/filepath"
	"strings"
)

// Find returns the Hyprglass source checkout directory.
// injectedByBuild is the compile-time -ldflags value; pass empty string if not available.
func Find(injectedByBuild string) string {
	var candidates []string
	add := func(p string) {
		if p != "" {
			candidates = append(candidates, p)
		}
	}
	add(os.Getenv("HYPRGLASS_SOURCE_ROOT"))
	add(injectedByBuild)
	if home, err := os.UserHomeDir(); err == nil {
		if b, err := os.ReadFile(filepath.Join(home, ".config", "hyprglass", "source-root")); err == nil {
			add(strings.TrimSpace(string(b)))
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		for {
			add(cwd)
			parent := filepath.Dir(cwd)
			if parent == cwd {
				break
			}
			cwd = parent
		}
	}
	if exe, err := os.Executable(); err == nil {
		add(filepath.Dir(exe))
	}
	for _, c := range candidates {
		if IsRoot(c) {
			return c
		}
	}
	return ""
}

// IsRoot reports whether dir looks like a Hyprglass source checkout.
func IsRoot(dir string) bool {
	if dir == "" {
		return false
	}
	for _, p := range []string{"install.sh", "cmd/hyprglass/main.go"} {
		if _, err := os.Stat(filepath.Join(dir, p)); err != nil {
			return false
		}
	}
	return true
}
