package service

import (
	"encoding/json"
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
		RepoScanDepth:            6,
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
