package globals

import (
	"reflect"
	"slices"
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

func TestCollectManifestFlagsNewFormats(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	files := map[string]string{
		"bun.lock":         `{}`,
		"CMakeLists.txt":   "find_package(Qt6 REQUIRED)\n",
		"deno.json":        `{}`,
		"go.work":          "go 1.21\n",
		"composer.json":    `{}`,
		"Gemfile":          "gem 'rails'\n",
		"pubspec.yaml":     "name: demo\n",
	}
	subfiles := make([]string, 0, len(files))
	for name, content := range files {
		testutil.WriteFile(t, root+"/"+name, content)
		subfiles = append(subfiles, name)
	}

	got, err := CollectManifestFlags(root, subfiles)
	if err != nil {
		t.Fatalf("CollectManifestFlags() error: %v", err)
	}

	want := []string{"@bun", "@cmake", "@deno", "@go-workspace", "@composer", "@ruby", "@dart"}
	if len(got) != len(want) {
		t.Fatalf("CollectManifestFlags() = %#v, want %d tags", got, len(want))
	}
	for _, tag := range want {
		if !slices.Contains(got, tag) {
			t.Fatalf("CollectManifestFlags() missing %q in %v", tag, got)
		}
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

func TestCollectSiblingProjectFlagsWithIdentifiers(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	files := map[string]string{
		"GNUmakefile":         "build:\n",
		"Dockerfile":          "FROM alpine\n",
		"docker-compose.yaml": "services: {}\n",
		"vite.config.mts":     "export default {}\n",
		"biome.jsonc":         "{}\n",
		"package.json":        `{}`,
	}
	subfiles := make([]string, 0, len(files))
	for name, content := range files {
		testutil.WriteFile(t, root+"/"+name, content)
		subfiles = append(subfiles, name)
	}

	got, err := CollectSiblingProjectFlags(root, subfiles, nil)
	if err != nil {
		t.Fatalf("CollectSiblingProjectFlags() error: %v", err)
	}

	want := []string{"@npm", "@makefile", "@docker", "@docker-compose", "@vite", "@biome"}
	if len(got) != len(want) {
		t.Fatalf("CollectSiblingProjectFlags() = %#v, want %d tags", got, len(want))
	}
	for _, tag := range want {
		if !slices.Contains(got, tag) {
			t.Fatalf("CollectSiblingProjectFlags() missing %q in %v", tag, got)
		}
	}
}
