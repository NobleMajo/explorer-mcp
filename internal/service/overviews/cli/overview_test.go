package cli

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestCommonCLIToolNamesShape(t *testing.T) {
	t.Parallel()

	if len(globals.CommonCLIToolNames) == 0 {
		t.Fatal("expected CommonCLIToolNames entries")
	}

	seen := make(map[string]struct{}, len(globals.CommonCLIToolNames))
	for _, name := range globals.CommonCLIToolNames {
		if name == "" {
			t.Fatal("expected non-empty cli tool name")
		}
		if _, ok := seen[name]; ok {
			t.Fatalf("duplicate cli tool name %q", name)
		}
		seen[name] = struct{}{}
	}
}

func TestCLIOverviewFindsToolsInPath(t *testing.T) {
	binDir := t.TempDir()
	for _, name := range []string{"go", "git", "make", "missing-tool"} {
		if name == "missing-tool" {
			continue
		}
		path := filepath.Join(binDir, name)
		testutil.WriteFile(t, path, "#!/bin/sh\n")
		if err := os.Chmod(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", binDir)

	result, err := CLIOverview()(false)
	if err != nil {
		t.Fatalf("CLIOverview() error: %v", err)
	}

	resp := result.(cliOverviewResponse)
	want := []string{"git", "go", "make"}
	if !slices.Equal(resp.CommonCliToolsFound, want) {
		t.Fatalf("commonCliToolsFound = %v, want %v", resp.CommonCliToolsFound, want)
	}
}

func TestCLIOverviewEmptyWhenNoToolsInPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	result, err := CLIOverview()(false)
	if err != nil {
		t.Fatalf("CLIOverview() error: %v", err)
	}

	resp := result.(cliOverviewResponse)
	if len(resp.CommonCliToolsFound) != 0 {
		t.Fatalf("commonCliToolsFound = %v, want empty", resp.CommonCliToolsFound)
	}
}
