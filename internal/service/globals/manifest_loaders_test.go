package globals

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestManifestLoadersShape(t *testing.T) {
	t.Parallel()

	want := []string{
		"Cargo.toml",
		"go.mod",
		"package.json",
		"pyproject.toml",
		"requirements.txt",
	}

	if len(ManifestLoaders) != len(want) {
		t.Fatalf("len(ManifestLoaders) = %d, want %d", len(ManifestLoaders), len(want))
	}

	for _, fileName := range want {
		if ManifestLoaders[fileName] == nil {
			t.Fatalf("ManifestLoaders missing %q", fileName)
		}
		if ManifestLoaderTags[fileName] == "" {
			t.Fatalf("ManifestLoaderTags missing %q", fileName)
		}
	}
}

func TestManifestLoadersInvokeAll(t *testing.T) {
	root := t.TempDir()

	fixtures := map[string]string{
		"go.mod":           "module demo\n\nrequire github.com/foo/bar v1.0.0\n",
		"package.json":     `{"dependencies":{"left-pad":"1.0.0"}}`,
		"requirements.txt": "flask>=3.0.0\n",
		"Cargo.toml":       "[package]\nname = \"demo\"\n",
		"pyproject.toml":   "[project]\nname = \"demo\"\n",
	}

	expectedTags := map[string]string{
		"go.mod":           "@go",
		"package.json":     "@npm",
		"requirements.txt": "@pip",
		"Cargo.toml":       "@cargo",
		"pyproject.toml":   "@python",
	}

	expected := map[string][]string{
		"go.mod":           {"github.com/foo/bar@v1.0.0 @direct"},
		"package.json":     {"left-pad@1.0.0 production"},
		"requirements.txt": {"flask@>=3.0.0"},
		"Cargo.toml":       nil,
		"pyproject.toml":   nil,
	}

	fileNames := make([]string, 0, len(ManifestLoaders))
	for fileName := range ManifestLoaders {
		fileNames = append(fileNames, fileName)
	}
	slices.Sort(fileNames)

	for _, fileName := range fileNames {
		content, ok := fixtures[fileName]
		if !ok {
			t.Fatalf("missing fixture for %q", fileName)
		}
		path := filepath.Join(root, fileName)
		testutil.WriteFile(t, path, content)

		loader := ManifestLoaders[fileName]
		if loader == nil {
			t.Fatalf("ManifestLoaders[%q] is nil", fileName)
		}

		tag, got, loaded, err := loader(root, path)
		if err != nil {
			t.Fatalf("ManifestLoaders[%q]() error: %v", fileName, err)
		}
		if !loaded {
			t.Fatalf("ManifestLoaders[%q]() loaded = false, want true", fileName)
		}
		if tag != expectedTags[fileName] {
			t.Fatalf("ManifestLoaders[%q]() tag = %q, want %q", fileName, tag, expectedTags[fileName])
		}

		want := expected[fileName]
		if len(got) != len(want) {
			t.Fatalf("%q dependencies = %v, want %v", fileName, got, want)
		}
		for _, entry := range want {
			if !slices.Contains(got, entry) {
				t.Fatalf("%q missing dependency %q in %v", fileName, entry, got)
			}
		}
	}
}
