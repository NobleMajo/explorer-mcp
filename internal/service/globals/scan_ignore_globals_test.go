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
		"__pycache__",
		".pytest_cache",
		".mypy_cache",
		".ruff_cache",
		".tox",
		".nox",
		".turbo",
		".sass-cache",
		".gradle",
		"htmlcov",
		".eslintcache",
		".stylelintcache",
		".tsbuildinfo",
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

	for _, name := range []string{"main.go", "src", ".gitignore", "README.md", "node_modules", "vendor"} {
		if IsScanIgnored(name) {
			t.Fatalf("IsScanIgnored(%q) = true, want false", name)
		}
	}
}
