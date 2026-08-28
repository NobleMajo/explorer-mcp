package workspacescan

import (
	"os"
	"slices"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestWalkTreeDepthZero(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	testutil.WriteFile(t, root+"/main.go", "package main\n")

	entries, err := WalkTree(root, StructureWalkOptions(root, 0, false, false))
	if err != nil {
		t.Fatalf("WalkTree() error: %v", err)
	}
	if entries != nil {
		t.Fatalf("WalkTree() entries = %#v, want nil", entries)
	}
}

func TestWalkTreeSortOrderAndFilePaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(root+"/adir", 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, root+"/zfile.go", "package z\n")
	testutil.WriteFile(t, root+"/adir/afile.go", "package a\n")
	testutil.WriteFile(t, root+"/afile.go", "package a\n")

	entries, err := WalkTree(root, StructureWalkOptions(root, 7, false, false))
	if err != nil {
		t.Fatalf("WalkTree() error: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 file entries, got %d: %+v", len(entries), entries)
	}

	fileNames := make([]string, 0, len(entries))
	for _, path := range entries {
		if path[len(path)-1] == '/' {
			t.Fatalf("unexpected directory path %q", path)
		}
		if i := lastSlash(path); i >= 0 {
			fileNames = append(fileNames, path[i+1:])
		} else {
			fileNames = append(fileNames, path)
		}
	}
	if !slices.IsSorted(fileNames) {
		t.Fatalf("files not sorted: %v", fileNames)
	}
}

func TestWalkTreeRespectsMaxDepth(t *testing.T) {
	t.Parallel()

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

	entries, err := WalkTree(root, StructureWalkOptions(root, maxDepth, false, false))
	if err != nil {
		t.Fatalf("WalkTree() error: %v", err)
	}

	gotMaxDepth := 0
	for _, path := range entries {
		if len(path) >= 3 && path[len(path)-3:] == "/**" {
			continue
		}
		depth := slashCount(path) + 1
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

func TestWalkTreeStopsOnNestedProjectFlags(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	testutil.WriteFile(t, root+"/main.go", "package main\n")
	if err := os.MkdirAll(root+"/packages/sub/nested", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(root+"/packages/sub/.git", 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, root+"/packages/sub/nested/file.go", "package nested\n")

	entries, err := WalkTree(root, StructureWalkOptions(root, 7, false, false))
	if err != nil {
		t.Fatalf("WalkTree() error: %v", err)
	}

	want := "packages/sub @git"
	if !slices.Contains(entries, want) {
		t.Fatalf("expected %q in entries, got %v", want, entries)
	}
	for _, forbidden := range []string{"packages/sub/nested/file.go", "packages/sub/nested"} {
		if slices.Contains(entries, forbidden) {
			t.Fatalf("expected scan to stop at flagged dir, found %q in %v", forbidden, entries)
		}
	}
}

func TestWalkTreeCollapsesOutDirAtDepthLimit(t *testing.T) {
	t.Parallel()

	const maxDepth = 2
	root := t.TempDir()
	if err := os.MkdirAll(root+"/src/dist/nested", 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, root+"/src/dist/nested/build.js", "x\n")
	testutil.WriteFile(t, root+"/src/app.go", "package app\n")

	entries, err := WalkTree(root, StructureWalkOptions(root, maxDepth, false, false))
	if err != nil {
		t.Fatalf("WalkTree() error: %v", err)
	}

	if !slices.Contains(entries, "src/app.go") {
		t.Fatalf("expected src/app.go, got %v", entries)
	}
	if !slices.Contains(entries, "src/dist/**") {
		t.Fatalf("expected nested out dir collapsed at depth limit, got %v", entries)
	}
	if slices.Contains(entries, "src/**") {
		t.Fatalf("expected src expanded until out dir, got %v", entries)
	}
}

func TestWalkTreeMissingDir(t *testing.T) {
	t.Parallel()

	_, err := WalkTree("/does/not/exist", StructureWalkOptions("/does/not/exist", 7, false, false))
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
}

func lastSlash(path string) int {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return i
		}
	}
	return -1
}

func slashCount(path string) int {
	count := 0
	for _, ch := range path {
		if ch == '/' {
			count++
		}
	}
	return count
}
