package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	filePath := filepath.Join(root, "demo.txt")
	if FileExists(filePath) {
		t.Fatal("expected missing file to return false")
	}
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !FileExists(filePath) {
		t.Fatal("expected existing file to return true")
	}
	if FileExists(root) {
		t.Fatal("expected directory to return false for FileExists")
	}
}

func TestDirExists(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if !DirExists(root) {
		t.Fatal("expected directory to return true")
	}
	filePath := filepath.Join(root, "demo.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if DirExists(filePath) {
		t.Fatal("expected file to return false for DirExists")
	}
}
