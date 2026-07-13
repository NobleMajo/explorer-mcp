package globals

import (
	"slices"
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
	if len(KnownContainerCLINames) != 8 {
		t.Fatalf("len(KnownContainerCLINames) = %d, want 8", len(KnownContainerCLINames))
	}
	if len(KnownContainerRuntimeCLINames) != 3 {
		t.Fatalf("len(KnownContainerRuntimeCLINames) = %d, want 3", len(KnownContainerRuntimeCLINames))
	}
	for _, runtimeName := range KnownContainerRuntimeCLINames {
		if !slices.Contains(KnownContainerCLINames, runtimeName) {
			t.Fatalf("runtime %q missing from KnownContainerCLINames", runtimeName)
		}
	}
}
