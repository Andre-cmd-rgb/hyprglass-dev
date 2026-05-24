package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopy(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src.txt")
	if err := os.WriteFile(src, []byte("hello hyprglass"), 0o644); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(t.TempDir(), "sub", "dst.txt")
	if err := Copy(src, dst); err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(got) != "hello hyprglass" {
		t.Fatalf("content mismatch: %q", got)
	}
}

func TestCopyMissingSource(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "dst.txt")
	if err := Copy("/nonexistent/file.txt", dst); err == nil {
		t.Fatal("expected error for missing source")
	}
}
