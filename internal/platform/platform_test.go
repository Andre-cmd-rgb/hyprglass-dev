package platform

import "testing"

func TestParseCachyOS(t *testing.T) {
	info := ParseOSRelease(`NAME="CachyOS"
ID=cachyos
ID_LIKE="arch"
PRETTY_NAME="CachyOS Linux"
`)
	if !info.CachyOS || !info.ArchLike {
		t.Fatalf("expected CachyOS arch-like, got %+v", info)
	}
	if got := info.PackageFile(); got != "packages/cachyos-core.txt" {
		t.Fatalf("PackageFile=%s", got)
	}
}

func TestParseArch(t *testing.T) {
	info := ParseOSRelease(`NAME="Arch Linux"
ID=arch
PRETTY_NAME="Arch Linux"
`)
	if info.CachyOS || !info.ArchLike {
		t.Fatalf("expected Arch-like non-CachyOS, got %+v", info)
	}
}
