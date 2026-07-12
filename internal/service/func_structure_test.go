package service

import (
	"os"
	"slices"
	"testing"
)

func TestRepoStructureSkipsIgnoredEntries(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root+"/main.go", "package main\n")
	writeFile(t, root+"/.gitignore", "*\n")
	if err := os.MkdirAll(root+"/node_modules/pkg", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(root+"/src", 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, root+"/src/app.go", "package src\n")

	chdir(t, root)

	jsonText, err := RepoStructure()
	if err != nil {
		t.Fatalf("RepoStructure() error: %v", err)
	}

	var resp repoStructureResponse
	parseJSONResponse(t, jsonText, &resp)

	if resp.ToolName != "repo_structure" || resp.MaxDepth != structureScanMaxDepth {
		t.Fatalf("unexpected meta: %+v", resp.responseMeta)
	}

	names := entryNames(resp.Entries)
	for _, forbidden := range []string{"node_modules", ".gitignore", ".dockerignore"} {
		if slices.Contains(names, forbidden) {
			t.Fatalf("expected %q to be ignored, entries=%v", forbidden, names)
		}
	}
	for _, required := range []string{"main.go", "src", "app.go"} {
		if !slices.Contains(names, required) {
			t.Fatalf("expected %q in entries, got %v", required, names)
		}
	}
}

func TestAppendStructureEntriesRespectsMaxDepth(t *testing.T) {
	root := t.TempDir()
	deep := root
	for i := 0; i < structureScanMaxDepth+2; i++ {
		deep = deep + "/level"
		if err := os.MkdirAll(deep, 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, deep+"/file.go", "package x\n")
	}

	entries := make([]structureEntry, 0)
	if err := appendStructureEntries(root, root, 0, &entries); err != nil {
		t.Fatalf("appendStructureEntries() error: %v", err)
	}

	maxDepth := 0
	for _, entry := range entries {
		if entry.Depth > maxDepth {
			maxDepth = entry.Depth
		}
	}
	if maxDepth > structureScanMaxDepth {
		t.Fatalf("max entry depth = %d, want <= %d", maxDepth, structureScanMaxDepth)
	}
}

func TestAppendStructureEntriesMissingDir(t *testing.T) {
	entries := make([]structureEntry, 0)
	err := appendStructureEntries(t.TempDir(), "/does/not/exist", 0, &entries)
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
}

func entryNames(entries []structureEntry) []string {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.EntryName)
	}
	return names
}
