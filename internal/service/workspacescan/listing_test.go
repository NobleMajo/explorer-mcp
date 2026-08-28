package workspacescan

import (
	"os"
	"reflect"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestSplitEntries(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(root+"/adir", 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, root+"/bfile.go", "package b\n")
	testutil.WriteFile(t, root+"/afile.go", "package a\n")

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}

	subfiles, subdirs := SplitEntries(entries)
	if !reflect.DeepEqual(subfiles, []string{"afile.go", "bfile.go"}) {
		t.Fatalf("SplitEntries() files = %#v", subfiles)
	}
	if !reflect.DeepEqual(subdirs, []string{"adir"}) {
		t.Fatalf("SplitEntries() dirs = %#v", subdirs)
	}
}

func TestReadListing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(root+"/child", 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, root+"/main.go", "package main\n")

	subfiles, subdirs, err := ReadListing(root)
	if err != nil {
		t.Fatalf("ReadListing() error: %v", err)
	}
	if !reflect.DeepEqual(subfiles, []string{"main.go"}) {
		t.Fatalf("ReadListing() files = %#v", subfiles)
	}
	if !reflect.DeepEqual(subdirs, []string{"child"}) {
		t.Fatalf("ReadListing() dirs = %#v", subdirs)
	}
}
