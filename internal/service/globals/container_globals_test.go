package globals

import (
	"strings"
	"testing"
)

func TestKnownContainerGlobalsShape(t *testing.T) {
	t.Parallel()

	if len(KnownContainerFileNames) != 22 {
		t.Fatalf("len(KnownContainerFileNames) = %d, want 22", len(KnownContainerFileNames))
	}
	if len(KnownContainerDirectoryNames) != 4 {
		t.Fatalf("len(KnownContainerDirectoryNames) = %d, want 4", len(KnownContainerDirectoryNames))
	}

	globPatterns := 0
	for _, name := range KnownContainerFileNames {
		if strings.Contains(name, "*") {
			globPatterns++
		}
	}
	if globPatterns != 8 {
		t.Fatalf("glob pattern count = %d, want 8", globPatterns)
	}
}
