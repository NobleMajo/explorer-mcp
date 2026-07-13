package globals

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestCollectManifestDependenciesGoToolDeps(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "go.mod"), `module demo

require golang.org/x/tools v0.30.0

tool golang.org/x/tools/cmd/goimports
`)

	got, err := CollectManifestDependencies(root, ManifestDepsSettings{ShowGoToolDeps: true})
	if err != nil {
		t.Fatalf("CollectManifestDependencies() error: %v", err)
	}

	for _, want := range []string{
		"golang.org/x/tools@v0.30.0 direct",
		"golang.org/x/tools/cmd/goimports@v0.30.0 tool",
	} {
		if !slices.Contains(got, want) {
			t.Fatalf("missing dependency %q, got %v", want, got)
		}
	}
}

func TestCollectManifestDependenciesGoToolDepsHidden(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "go.mod"), `module demo

require golang.org/x/tools v0.30.0

tool golang.org/x/tools/cmd/goimports
`)

	got, err := CollectManifestDependencies(root, ManifestDepsSettings{ShowGoToolDeps: false})
	if err != nil {
		t.Fatalf("CollectManifestDependencies() error: %v", err)
	}

	if len(got) != 1 || got[0] != "golang.org/x/tools@v0.30.0 direct" {
		t.Fatalf("unexpected dependencies: %v", got)
	}
	for _, entry := range got {
		if strings.HasSuffix(entry, " tool") {
			t.Fatalf("expected tool deps hidden, got %v", got)
		}
	}
}

func TestCollectManifestDependenciesMergesAllManifests(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "go.mod"), "module demo\n\nrequire github.com/foo/bar v1.0.0\n")
	testutil.WriteFile(t, filepath.Join(root, "package.json"), `{"dependencies":{"left-pad":"1.0.0"}}`)
	testutil.WriteFile(t, filepath.Join(root, "bun.lock"), `{
  "workspaces": { "": { "dependencies": { "zod": "^4.1.12" } } },
  "packages": { "zod": ["zod@4.1.12", "", {}, "hash"] }
}`)
	testutil.WriteFile(t, filepath.Join(root, "CMakeLists.txt"), "find_package(Qt6 REQUIRED)\n")
	testutil.WriteFile(t, filepath.Join(root, "composer.json"), `{"require":{"php":"^8.1"}}`)

	got, err := CollectManifestDependencies(root, DefaultManifestDepsSettings())
	if err != nil {
		t.Fatalf("CollectManifestDependencies() error: %v", err)
	}

	for _, want := range []string{
		"github.com/foo/bar@v1.0.0 direct",
		"left-pad@1.0.0 direct",
		"zod@^4.1.12 direct",
		"Qt6 direct",
		"php@^8.1 direct",
	} {
		if !slices.Contains(got, want) {
			t.Fatalf("missing dependency %q, got %v", want, got)
		}
	}
}

func TestDefaultManifestDepsSettings(t *testing.T) {
	t.Parallel()

	settings := DefaultManifestDepsSettings()
	if !settings.ShowGoToolDeps {
		t.Fatal("expected ShowGoToolDeps true by default")
	}
}

func TestCollectManifestDependenciesUsesStandardScopes(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "go.mod"), `module demo

require (
	github.com/foo/bar v1.0.0
	github.com/indirect/dep v0.1.0 // indirect
)

tool golang.org/x/tools/cmd/goimports
`)
	testutil.WriteFile(t, filepath.Join(root, "package.json"), `{"dependencies":{"alpha":"1.0.0"},"devDependencies":{"eslint":"9.0.0"}}`)
	testutil.WriteFile(t, filepath.Join(root, "bun.lock"), `{
  "workspaces": {
    "": {
      "dependencies": { "zod": "^4.1.12" },
      "devDependencies": { "prettier": "3.6.2" }
    }
  },
  "packages": {
    "zod": ["zod@4.1.12", "", {}, "hash"],
    "left-pad": ["left-pad@1.0.0", "", {}, "hash"]
  }
}`)

	got, err := CollectManifestDependencies(root, DefaultManifestDepsSettings())
	if err != nil {
		t.Fatalf("CollectManifestDependencies() error: %v", err)
	}

	assertStandardDepScopes(t, got)

	for _, want := range []string{
		"github.com/foo/bar@v1.0.0 direct",
		"github.com/indirect/dep@v0.1.0 indirect",
		"golang.org/x/tools/cmd/goimports tool",
		"alpha@1.0.0 direct",
		"eslint@9.0.0 dev",
		"zod@^4.1.12 direct",
		"prettier@3.6.2 dev",
		"left-pad@1.0.0 indirect",
	} {
		if !slices.Contains(got, want) {
			t.Fatalf("missing dependency %q, got %v", want, got)
		}
	}
}
