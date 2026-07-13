package structure

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestRepoStructureSkipsIgnoredEntries(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/main.go", "package main\n")
	testutil.WriteFile(t, root+"/.gitignore", "ignored-output/\n*.tmp\n")
	testutil.WriteFile(t, root+"/.dockerignore", "container-artifacts/\n")
	for _, dirName := range globals.ScanIgnoreFiles {
		if err := os.MkdirAll(root+"/"+dirName, 0o755); err != nil {
			t.Fatal(err)
		}
		testutil.WriteFile(t, root+"/"+dirName+"/ignored.txt", "x\n")
	}
	if err := os.MkdirAll(root+"/src", 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, root+"/src/app.go", "package src\n")
	if err := os.MkdirAll(root+"/ignored-output", 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, root+"/ignored-output/skip.txt", "x\n")
	testutil.WriteFile(t, root+"/notes.tmp", "x\n")

	testutil.Chdir(t, root)

	result, err := StructureOverview()(false)
	if err != nil {
		t.Fatalf("StructureOverview() error: %v", err)
	}

	resp, ok := result.(repoStructureResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	if resp.ToolName != "repo_structure" {
		t.Fatalf("unexpected meta: %+v", resp.Meta)
	}
	if resp.EntryCount != len(resp.Entries) {
		t.Fatalf("entryCount = %d, len(entries) = %d", resp.EntryCount, len(resp.Entries))
	}

	names := entryNames(resp.Entries)
	for _, forbidden := range append(slices.Clone(globals.ScanIgnoreFiles), "ignored.txt", "ignored-output", "skip.txt", "notes.tmp") {
		if slices.Contains(names, forbidden) {
			t.Fatalf("expected %q to be ignored, entries=%v", forbidden, names)
		}
	}
	for _, required := range []string{"main.go", "app.go"} {
		if !slices.Contains(names, required) {
			t.Fatalf("expected %q in entries, got %v", required, names)
		}
	}
	for _, entry := range resp.Entries {
		if entry.IsDirectory {
			t.Fatalf("expected files only, got directory entry %+v", entry)
		}
		if strings.HasSuffix(entry.RelativePath, "/") {
			t.Fatalf("expected file path, got directory path %q", entry.RelativePath)
		}
	}
}

func TestStructureSortOrderAndFilePaths(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(root+"/adir", 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, root+"/zfile.go", "package z\n")
	testutil.WriteFile(t, root+"/adir/afile.go", "package a\n")
	testutil.WriteFile(t, root+"/afile.go", "package a\n")

	entries := make([]structureEntry, 0)
	if err := appendStructureEntries(root, root, 0, &entries, newScanState()); err != nil {
		t.Fatalf("appendStructureEntries() error: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 file entries, got %d: %+v", len(entries), entries)
	}

	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDirectory {
			t.Fatalf("expected files only, got directory %+v", entry)
		}
		if strings.Contains(entry.RelativePath, "/") && strings.HasSuffix(entry.RelativePath, "/") {
			t.Fatalf("unexpected directory path %q", entry.RelativePath)
		}
		fileNames = append(fileNames, entry.EntryName)
	}
	if !slices.IsSorted(fileNames) {
		t.Fatalf("files not sorted: %v", fileNames)
	}
}

func TestAppendStructureEntriesRespectsMaxDepth(t *testing.T) {
	root := t.TempDir()
	deep := root
	for i := 0; i < globals.StructureScanMaxDepth+2; i++ {
		deep = deep + "/level"
		if err := os.MkdirAll(deep, 0o755); err != nil {
			t.Fatal(err)
		}
		testutil.WriteFile(t, deep+"/file.go", "package x\n")
	}

	entries := make([]structureEntry, 0)
	if err := appendStructureEntries(root, root, 0, &entries, newScanState()); err != nil {
		t.Fatalf("appendStructureEntries() error: %v", err)
	}

	maxDepth := 0
	for _, entry := range entries {
		if entry.Depth > maxDepth {
			maxDepth = entry.Depth
		}
	}
	if maxDepth > globals.StructureScanMaxDepth {
		t.Fatalf("max entry depth = %d, want <= %d", maxDepth, globals.StructureScanMaxDepth)
	}
}

func TestRepoStructureFollowsGitIgnore(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/main.go", "package main\n")
	testutil.WriteFile(t, root+"/.gitignore", "build/\n*.log\n")
	if err := os.MkdirAll(root+"/build", 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, root+"/build/out.go", "package build\n")
	testutil.WriteFile(t, root+"/app.log", "log\n")

	testutil.Chdir(t, root)

	result, err := StructureOverview()(false)
	if err != nil {
		t.Fatalf("StructureOverview() error: %v", err)
	}

	resp := result.(repoStructureResponse)
	names := entryNames(resp.Entries)
	for _, forbidden := range []string{"build", "out.go", "app.log"} {
		if slices.Contains(names, forbidden) {
			t.Fatalf("expected %q to be gitignored, entries=%v", forbidden, names)
		}
	}
	if !slices.Contains(names, "main.go") {
		t.Fatalf("expected main.go in entries, got %v", names)
	}
}

func TestRepoStructureNestedGitIgnore(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/main.go", "package main\n")
	if err := os.MkdirAll(root+"/pkg", 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, root+"/pkg/.gitignore", "generated/\n")
	testutil.WriteFile(t, root+"/pkg/manual.go", "package pkg\n")
	if err := os.MkdirAll(root+"/pkg/generated", 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, root+"/pkg/generated/skip.go", "package generated\n")

	testutil.Chdir(t, root)

	result, err := StructureOverview()(false)
	if err != nil {
		t.Fatalf("StructureOverview() error: %v", err)
	}

	resp := result.(repoStructureResponse)
	names := entryNames(resp.Entries)
	for _, forbidden := range []string{"generated", "skip.go"} {
		if slices.Contains(names, forbidden) {
			t.Fatalf("expected %q to be gitignored, entries=%v", forbidden, names)
		}
	}
	if !slices.Contains(names, "manual.go") {
		t.Fatalf("expected manual.go in entries, got %v", names)
	}
	if slices.Contains(names, "pkg") {
		t.Fatalf("expected no directory entries, got %v", names)
	}
}

func TestRepoStructureFollowGitIgnoreDisabled(t *testing.T) {
	old := globals.FollowGitIgnore
	globals.FollowGitIgnore = false
	t.Cleanup(func() { globals.FollowGitIgnore = old })

	root := t.TempDir()
	testutil.WriteFile(t, root+"/main.go", "package main\n")
	testutil.WriteFile(t, root+"/.gitignore", "build/\n")
	if err := os.MkdirAll(root+"/build", 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, root+"/build/out.go", "package build\n")

	testutil.Chdir(t, root)

	result, err := StructureOverview()(false)
	if err != nil {
		t.Fatalf("StructureOverview() error: %v", err)
	}

	resp := result.(repoStructureResponse)

	names := entryNames(resp.Entries)
	if !slices.Contains(names, "out.go") {
		t.Fatalf("expected out.go when gitignore disabled, entries=%v", names)
	}
	if slices.Contains(names, "build") {
		t.Fatalf("expected no directory entries, got %v", names)
	}
}

func TestRepoStructureIncludesEnvFiles(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/.env", "PORT=8080\n")
	testutil.WriteFile(t, root+"/.env.project", "PROJECT_VERSION=1.0.0\n")
	testutil.Chdir(t, root)

	result, err := StructureOverview()(false)
	if err != nil {
		t.Fatalf("StructureOverview() error: %v", err)
	}

	resp := result.(repoStructureResponse)
	paths := make([]string, 0, len(resp.Entries))
	for _, entry := range resp.Entries {
		paths = append(paths, entry.RelativePath)
	}
	for _, want := range []string{".env", ".env.project"} {
		if !slices.Contains(paths, want) {
			t.Fatalf("expected %q in entries, got %v", want, paths)
		}
	}
}

func TestAppendStructureEntriesSkipsIgnoredFiles(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/main.go", "package main\n")
	for _, fileName := range globals.IgnoreFiles {
		testutil.WriteFile(t, root+"/"+fileName, "# ignore rules\n")
	}

	entries := make([]structureEntry, 0)
	if err := appendStructureEntries(root, root, 0, &entries, newScanState()); err != nil {
		t.Fatalf("appendStructureEntries() error: %v", err)
	}

	names := entryNames(entries)
	for _, fileName := range globals.IgnoreFiles {
		if slices.Contains(names, fileName) {
			t.Fatalf("expected %q to be skipped, entries=%v", fileName, names)
		}
	}
	if !slices.Contains(names, "main.go") {
		t.Fatalf("expected main.go in entries, got %v", names)
	}
}

func TestAppendStructureEntriesMissingDir(t *testing.T) {
	entries := make([]structureEntry, 0)
	err := appendStructureEntries(t.TempDir(), "/does/not/exist", 0, &entries, newScanState())
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
