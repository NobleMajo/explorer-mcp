package workspacescan

import (
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestWalkDownStopsOnFlags(t *testing.T) {
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

	opts := Options{
		Root:             root,
		Depth:            5,
		CheckFlags:       true,
		StopOnFlags:      true,
		ShowNonFlag:      true,
		IgnoreHiddenDirs: true,
	}

	var visited []string
	err := WalkDown(root, opts, func(listing Listing) error {
		visited = append(visited, listing.RelPath)
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDown() error: %v", err)
	}

	if !slices.Contains(visited, "packages/sub") {
		t.Fatalf("expected packages/sub visited, got %v", visited)
	}
	if slices.Contains(visited, "packages/sub/nested") {
		t.Fatalf("expected no descent into flagged dir, got %v", visited)
	}
}

func TestWalkDownShowNonFlagOnly(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(root+"/plain/deep", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(root+"/flagged/.git", 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, root+"/plain/deep/file.go", "package deep\n")

	opts := Options{
		Root:             root,
		Depth:            3,
		CheckFlags:       true,
		StopOnFlags:      true,
		ShowNonFlag:      false,
		IgnoreHiddenDirs: true,
	}

	var visited []string
	err := WalkDown(root, opts, func(listing Listing) error {
		visited = append(visited, listing.RelPath)
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDown() error: %v", err)
	}

	want := []string{"flagged"}
	if !reflect.DeepEqual(visited, want) {
		t.Fatalf("WalkDown() visited = %v, want %v", visited, want)
	}
}

func TestWalkDownIgnoresPathsAndSkipDirs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(root+"/keep/child", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(root+"/skip/child", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(root+"/ignored/child", 0o755); err != nil {
		t.Fatal(err)
	}

	opts := Options{
		Root:             root,
		Depth:            2,
		ShowNonFlag:      true,
		IgnoreHiddenDirs: true,
		SkipDirs:         []string{"skip"},
		IgnorePaths:      []string{root + "/ignored"},
	}

	var visited []string
	err := WalkDown(root, opts, func(listing Listing) error {
		visited = append(visited, listing.RelPath)
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDown() error: %v", err)
	}

	for _, forbidden := range []string{"skip", "skip/child", "ignored", "ignored/child"} {
		if slices.Contains(visited, forbidden) {
			t.Fatalf("expected %q skipped, visited=%v", forbidden, visited)
		}
	}
	if !slices.Contains(visited, "keep") || !slices.Contains(visited, "keep/child") {
		t.Fatalf("expected keep dirs visited, got %v", visited)
	}
}

func TestWalkDownDepthZero(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(root+"/child", 0o755); err != nil {
		t.Fatal(err)
	}

	called := false
	err := WalkDown(root, Options{Root: root, Depth: 0, ShowNonFlag: true}, func(listing Listing) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDown() error: %v", err)
	}
	if called {
		t.Fatal("expected no traversal when depth < 1")
	}
}
