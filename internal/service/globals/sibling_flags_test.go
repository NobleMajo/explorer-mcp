package globals

import (
	"reflect"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestCollectManifestFlags(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	testutil.WriteFile(t, root+"/go.mod", "module demo\n")
	testutil.WriteFile(t, root+"/package.json", "{}\n")

	got, err := CollectManifestFlags(root, []string{"go.mod", "package.json", "README.md"})
	if err != nil {
		t.Fatalf("CollectManifestFlags() error: %v", err)
	}
	want := []string{"@go", "@npm"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CollectManifestFlags() = %#v, want %#v", got, want)
	}
}

func TestCollectManifestFlagsSkipsMissingFiles(t *testing.T) {
	t.Parallel()

	got, err := CollectManifestFlags(t.TempDir(), []string{"go.mod"})
	if err != nil {
		t.Fatalf("CollectManifestFlags() error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("CollectManifestFlags() = %#v, want empty", got)
	}
}

func TestCollectSiblingProjectFlags(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	testutil.WriteFile(t, root+"/go.mod", "module demo\n")
	testutil.WriteFile(t, root+"/Makefile", "build:\n")

	got, err := CollectSiblingProjectFlags(root, []string{"go.mod", "Makefile"}, []string{".git"})
	if err != nil {
		t.Fatalf("CollectSiblingProjectFlags() error: %v", err)
	}
	want := []string{"@git", "@go", "@makefile"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CollectSiblingProjectFlags() = %#v, want %#v", got, want)
	}
}
