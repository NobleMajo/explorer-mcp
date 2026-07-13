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

	expected := map[string]struct {
		ecosystem string
		parsed    bool
	}{
		"go.mod":           {ecosystem: "go", parsed: true},
		"package.json":     {ecosystem: "node", parsed: true},
		"requirements.txt": {ecosystem: "python", parsed: true},
		"Cargo.toml":       {ecosystem: "rust", parsed: false},
		"pyproject.toml":   {ecosystem: "python", parsed: false},
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

		got, err := loader(root, path)
		if err != nil {
			t.Fatalf("ManifestLoaders[%q]() error: %v", fileName, err)
		}

		want := expected[fileName]
		if got.EcosystemName != want.ecosystem {
			t.Fatalf("%q ecosystem = %q, want %q", fileName, got.EcosystemName, want.ecosystem)
		}
		if got.ManifestFilePath != fileName {
			t.Fatalf("%q manifest path = %q, want %q", fileName, got.ManifestFilePath, fileName)
		}
		if !got.ManifestFileExists {
			t.Fatalf("%q manifestFileExists = false", fileName)
		}
		if got.IsParsed != want.parsed {
			t.Fatalf("%q isParsed = %v, want %v", fileName, got.IsParsed, want.parsed)
		}
	}
}
