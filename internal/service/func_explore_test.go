package service

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestExploreCombinesToolSections(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root+"/main.go", "package main\n")
	writeFile(t, root+"/go.mod", "module demo\n")

	chdir(t, root)

	jsonText, err := Explore()
	if err != nil {
		t.Fatalf("Explore() error: %v", err)
	}

	var resp exploreResponse
	parseJSONResponse(t, jsonText, &resp)

	if resp.ToolName != "explore" {
		t.Fatalf("toolName = %q", resp.ToolName)
	}

	assertSectionToolName(t, "repoStructure", resp.RepoStructure, "repo_structure")
	assertSectionToolName(t, "gitOverview", resp.GitOverview, "git_overview")
	assertSectionToolName(t, "workspaceContext", resp.WorkspaceContext, "workspace_context")
	assertSectionToolName(t, "dependencies", resp.Dependencies, "dependencies")
}

func TestToolSectionEmbedsJSON(t *testing.T) {
	t.Parallel()

	section, err := toolSection(func() (string, error) {
		return marshalResponse(responseMeta{
			ToolName:      "demo",
			SchemaVersion: schemaVersion,
		})
	})
	if err != nil {
		t.Fatalf("toolSection() error: %v", err)
	}

	var meta responseMeta
	if err := json.Unmarshal(section, &meta); err != nil {
		t.Fatalf("unmarshal section: %v", err)
	}
	if meta.ToolName != "demo" {
		t.Fatalf("toolName = %q", meta.ToolName)
	}
}

func TestToolSectionPropagatesError(t *testing.T) {
	t.Parallel()

	_, err := toolSection(func() (string, error) {
		return "", errors.New("section failed")
	})
	if err == nil {
		t.Fatal("expected toolSection error")
	}
}

func assertSectionToolName(t *testing.T, field string, raw json.RawMessage, want string) {
	t.Helper()

	var meta responseMeta
	if err := json.Unmarshal(raw, &meta); err != nil {
		t.Fatalf("unmarshal %s: %v", field, err)
	}
	if meta.ToolName != want {
		t.Fatalf("%s toolName = %q, want %q", field, meta.ToolName, want)
	}
}
