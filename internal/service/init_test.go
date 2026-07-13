package service

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/config"
	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestDirectJsonResultUsesConfigSettings(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/main.go", "package main\n")
	testutil.Chdir(t, root)

	out, err := DirectJsonResult(&config.AppConfig{
		DisableStructureOverview: true,
		ProjectScanDepth:         6,
	})
	if err != nil {
		t.Fatalf("DirectJsonResult() error: %v", err)
	}

	var resp map[string]json.RawMessage
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := resp["structure"]; ok {
		t.Fatal("expected structure omitted from DirectJsonResult output")
	}
	if _, ok := resp["cli"]; ok {
		t.Fatal("expected cli omitted by default in DirectJsonResult output")
	}
}

func TestDirectJsonResultShowGoToolDepsConfig(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/go.mod", `module demo

require golang.org/x/tools v0.30.0

tool golang.org/x/tools/cmd/goimports
`)
	testutil.Chdir(t, root)

	out, err := DirectJsonResult(&config.AppConfig{
		ShowGoToolDeps: false,
	})
	if err != nil {
		t.Fatalf("DirectJsonResult() error: %v", err)
	}

	var resp struct {
		Dependencies []string `json:"dependencies"`
	}
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, entry := range resp.Dependencies {
		if strings.HasSuffix(entry, " tool") {
			t.Fatalf("expected tool deps hidden, got %v", resp.Dependencies)
		}
	}
	if !slices.Contains(resp.Dependencies, "golang.org/x/tools@v0.30.0 direct") {
		t.Fatalf("missing require dep, got %v", resp.Dependencies)
	}
}
