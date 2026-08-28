package workspacescan

import (
	"reflect"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestCollectFlagsDisabled(t *testing.T) {
	t.Parallel()

	flags, err := CollectFlags(Options{CheckFlags: false}, t.TempDir(), []string{"go.mod"}, []string{".git"})
	if err != nil {
		t.Fatalf("CollectFlags() error: %v", err)
	}
	if flags != nil {
		t.Fatalf("CollectFlags() = %#v, want nil", flags)
	}
}

func TestCollectFlagsGitRepo(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	testutil.WriteFile(t, root+"/go.mod", "module demo\n")

	flags, err := CollectFlags(Options{CheckFlags: true}, root, []string{"go.mod"}, []string{".git"})
	if err != nil {
		t.Fatalf("CollectFlags() error: %v", err)
	}
	want := []string{"@git", "@go"}
	if !reflect.DeepEqual(flags, want) {
		t.Fatalf("CollectFlags() = %#v, want %#v", flags, want)
	}
}

func TestHasFlags(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if !HasFlags(Options{CheckFlags: true}, root, nil, []string{".git"}) {
		t.Fatal("HasFlags() = false, want true for .git subdir")
	}
	if HasFlags(Options{CheckFlags: false}, root, nil, []string{".git"}) {
		t.Fatal("HasFlags() = true, want false when CheckFlags disabled")
	}
}

func TestFormatPathWithFlags(t *testing.T) {
	t.Parallel()

	got := FormatPathWithFlags("packages/sub", []string{"@git", "@go"})
	want := "packages/sub @git @go"
	if got != want {
		t.Fatalf("FormatPathWithFlags() = %q, want %q", got, want)
	}
	if FormatPathWithFlags("plain", nil) != "plain" {
		t.Fatal("FormatPathWithFlags() should return path unchanged when no flags")
	}
}
