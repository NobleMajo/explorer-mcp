package structure

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

const testRepoScanDepth = 7

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

	result, err := StructureOverview(testRepoScanDepth)()(false)
	if err != nil {
		t.Fatalf("StructureOverview() error: %v", err)
	}

	resp, ok := result.(repoStructureResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	if resp.EntryCount == nil || *resp.EntryCount != len(resp.Entries) {
		t.Fatalf("entryCount = %v, len(entries) = %d", resp.EntryCount, len(resp.Entries))
	}

	names := entryBaseNames(resp.Entries)
	for _, forbidden := range append(slices.Clone(globals.ScanIgnoreFiles), "ignored.txt") {
		if slices.Contains(names, forbidden) {
			t.Fatalf("expected %q to be ignored, entries=%v", forbidden, names)
		}
	}
	for _, required := range []string{"main.go", "app.go", "skip.txt", "notes.tmp"} {
		if !slices.Contains(names, required) {
			t.Fatalf("expected %q in entries, got %v", required, names)
		}
	}
	for _, path := range resp.Entries {
		if strings.HasSuffix(path, "/") {
			t.Fatalf("expected file path, got directory path %q", path)
		}
	}
}

func TestRepoStructureDepthZero(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/main.go", "package main\n")
	testutil.Chdir(t, root)

	result, err := StructureOverview(0)()(false)
	if err != nil {
		t.Fatalf("StructureOverview() error: %v", err)
	}

	resp := result.(repoStructureResponse)
	if resp.RepoScanDepthLimit == nil || *resp.RepoScanDepthLimit != 0 {
		t.Fatalf("repoScanDepthLimit = %v, want 0", resp.RepoScanDepthLimit)
	}
	if resp.EntryCount != nil || len(resp.Entries) != 0 {
		t.Fatalf("expected no entries, got %+v", resp)
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

	entries := make([]string, 0)
	if err := appendStructureEntries(root, root, 0, testRepoScanDepth, &entries); err != nil {
		t.Fatalf("appendStructureEntries() error: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 file entries, got %d: %+v", len(entries), entries)
	}

	fileNames := make([]string, 0, len(entries))
	for _, path := range entries {
		if strings.Contains(path, "/") && strings.HasSuffix(path, "/") {
			t.Fatalf("unexpected directory path %q", path)
		}
		fileNames = append(fileNames, filepathBaseName(path))
	}
	if !slices.IsSorted(fileNames) {
		t.Fatalf("files not sorted: %v", fileNames)
	}
}

func TestAppendStructureEntriesRespectsMaxDepth(t *testing.T) {
	const maxDepth = 3
	root := t.TempDir()
	deep := root
	for i := 0; i < maxDepth+2; i++ {
		deep = deep + "/level"
		if err := os.MkdirAll(deep, 0o755); err != nil {
			t.Fatal(err)
		}
		testutil.WriteFile(t, deep+"/file.go", "package x\n")
	}

	entries := make([]string, 0)
	if err := appendStructureEntries(root, root, 0, maxDepth, &entries); err != nil {
		t.Fatalf("appendStructureEntries() error: %v", err)
	}

	gotMaxDepth := 0
	for _, path := range entries {
		if strings.HasSuffix(path, "/**") {
			continue
		}
		depth := strings.Count(path, "/") + 1
		if depth > gotMaxDepth {
			gotMaxDepth = depth
		}
	}
	if gotMaxDepth > maxDepth {
		t.Fatalf("max entry depth = %d, want <= %d", gotMaxDepth, maxDepth)
	}
	if !slices.Contains(entries, "level/level/level/**") {
		t.Fatalf("expected depth truncation marker level/level/level/**, got %v", entries)
	}
}

func TestAppendStructureEntriesMarksDeeperDirectories(t *testing.T) {
	const maxDepth = 3
	root := t.TempDir()
	if err := os.MkdirAll(root+"/internal/service/overviews/structure", 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, root+"/internal/service/handlers.go", "package service\n")
	testutil.WriteFile(t, root+"/internal/service/overviews/structure/overview.go", "package structure\n")
	testutil.WriteFile(t, root+"/internal/config/config.go", "package config\n")

	entries := make([]string, 0)
	if err := appendStructureEntries(root, root, 0, maxDepth, &entries); err != nil {
		t.Fatalf("appendStructureEntries() error: %v", err)
	}

	for _, want := range []string{
		"internal/config/config.go",
		"internal/service/handlers.go",
		"internal/service/overviews/**",
	} {
		if !slices.Contains(entries, want) {
			t.Fatalf("expected %q in entries, got %v", want, entries)
		}
	}
	for _, path := range entries {
		if strings.HasSuffix(path, "/**") && strings.Count(strings.TrimSuffix(path, "/**"), "/")+1 > maxDepth {
			t.Fatalf("truncation marker beyond depth limit: %q", path)
		}
	}
}

func TestAppendStructureEntriesOmitsMarkerForEmptyDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(root+"/empty/nested", 0o755); err != nil {
		t.Fatal(err)
	}

	entries := make([]string, 0)
	if err := appendStructureEntries(root, root, 0, 2, &entries); err != nil {
		t.Fatalf("appendStructureEntries() error: %v", err)
	}

	for _, path := range entries {
		if strings.HasSuffix(path, "/**") {
			t.Fatalf("expected no truncation marker for empty dirs, got %q", path)
		}
	}
}

func TestRepoStructureIncludesScanDepthLimit(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/main.go", "package main\n")
	testutil.Chdir(t, root)

	const depth = 4
	result, err := StructureOverview(depth)()(false)
	if err != nil {
		t.Fatalf("StructureOverview() error: %v", err)
	}

	resp := result.(repoStructureResponse)
	if resp.RepoScanDepthLimit == nil || *resp.RepoScanDepthLimit != depth {
		t.Fatalf("repoScanDepthLimit = %v, want %d", resp.RepoScanDepthLimit, depth)
	}
}

func TestRepoStructureDoesNotFollowGitIgnore(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/main.go", "package main\n")
	testutil.WriteFile(t, root+"/.gitignore", "build/\n*.log\n")
	if err := os.MkdirAll(root+"/build", 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, root+"/build/out.go", "package build\n")
	testutil.WriteFile(t, root+"/app.log", "log\n")

	testutil.Chdir(t, root)

	result, err := StructureOverview(testRepoScanDepth)()(false)
	if err != nil {
		t.Fatalf("StructureOverview() error: %v", err)
	}

	resp := result.(repoStructureResponse)
	names := entryBaseNames(resp.Entries)
	for _, want := range []string{"main.go", "out.go", "app.log"} {
		if !slices.Contains(names, want) {
			t.Fatalf("expected %q in entries, got %v", want, names)
		}
	}
	if slices.Contains(names, ".gitignore") {
		t.Fatalf("expected .gitignore omitted, entries=%v", names)
	}
}

func TestRepoStructureIncludesEnvFiles(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/.env", "PORT=8080\n")
	testutil.WriteFile(t, root+"/.env.project", "PROJECT_VERSION=1.0.0\n")
	testutil.Chdir(t, root)

	result, err := StructureOverview(testRepoScanDepth)()(false)
	if err != nil {
		t.Fatalf("StructureOverview() error: %v", err)
	}

	resp := result.(repoStructureResponse)
	for _, want := range []string{".env", ".env.project"} {
		if !slices.Contains(resp.Entries, want) {
			t.Fatalf("expected %q in entries, got %v", want, resp.Entries)
		}
	}
}

func TestAppendStructureEntriesSkipsIgnoredFiles(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/main.go", "package main\n")
	for _, fileName := range globals.IgnoreFiles {
		testutil.WriteFile(t, root+"/"+fileName, "# ignore rules\n")
	}

	entries := make([]string, 0)
	if err := appendStructureEntries(root, root, 0, testRepoScanDepth, &entries); err != nil {
		t.Fatalf("appendStructureEntries() error: %v", err)
	}

	names := entryBaseNames(entries)
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
	entries := make([]string, 0)
	err := appendStructureEntries(t.TempDir(), "/does/not/exist", 0, testRepoScanDepth, &entries)
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
}

func entryBaseNames(entries []string) []string {
	names := make([]string, 0, len(entries))
	for _, path := range entries {
		names = append(names, filepathBaseName(path))
	}
	return names
}

func filepathBaseName(path string) string {
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	return path
}
