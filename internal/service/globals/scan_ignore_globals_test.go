package globals

import (
	"slices"
	"testing"
)

func TestScanIgnoreGlobalsShape(t *testing.T) {
	t.Parallel()

	want := []string{
		".tmp",
		"tmp",
		".git",
		".angular",
		".cache",
		"cache",
		"vendor",
		"node_modules",
	}

	if len(ScanIgnoreFiles) != len(want) {
		t.Fatalf("len(ScanIgnoreFiles) = %d, want %d", len(ScanIgnoreFiles), len(want))
	}

	for _, name := range want {
		if !slices.Contains(ScanIgnoreFiles, name) {
			t.Fatalf("ScanIgnoreFiles missing %q", name)
		}
	}
}

func TestIsScanIgnoredCoversAllGlobals(t *testing.T) {
	t.Parallel()

	for _, name := range ScanIgnoreFiles {
		if !IsScanIgnored(name) {
			t.Fatalf("IsScanIgnored(%q) = false, want true", name)
		}
	}

	for _, name := range []string{"main.go", "src", ".gitignore", "README.md"} {
		if IsScanIgnored(name) {
			t.Fatalf("IsScanIgnored(%q) = true, want false", name)
		}
	}
}

func TestIgnoreFilesGlobalsShape(t *testing.T) {
	t.Parallel()

	want := []string{".gitignore", ".dockerignore"}

	if len(IgnoreFiles) != len(want) {
		t.Fatalf("len(IgnoreFiles) = %d, want %d", len(IgnoreFiles), len(want))
	}

	for _, name := range want {
		if !slices.Contains(IgnoreFiles, name) {
			t.Fatalf("IgnoreFiles missing %q", name)
		}
	}
}

func TestIsIgnoredFileCoversAllGlobals(t *testing.T) {
	t.Parallel()

	for _, name := range IgnoreFiles {
		if !IsIgnoredFile(name) {
			t.Fatalf("IsIgnoredFile(%q) = false, want true", name)
		}
	}

	for _, name := range []string{"main.go", "Dockerfile", "go.mod", ".git"} {
		if IsIgnoredFile(name) {
			t.Fatalf("IsIgnoredFile(%q) = true, want false", name)
		}
	}
}
