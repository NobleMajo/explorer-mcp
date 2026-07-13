package globals

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestValidateRealWorkspaceManifests(t *testing.T) {
	cases := []struct {
		name     string
		path     string
		loader   func(root, path string) (string, []string, bool, error)
		tag      string
		contains []string
		minDeps  int
	}{
		{
			name:   "kyro bun.lock",
			path:   "/home/majo/ws/games/kyro/bun.lock",
			loader: LoadBunLockManifest,
			tag:    "@bun",
			contains: []string{
				"cprox@^1.9.39 direct",
				"zod@^4.1.12 direct",
				"prettier@3.6.2 dev",
				"typescript@^5 direct",
			},
			minDeps: 10,
		},
		{
			name:   "obsbot CMakeLists.txt",
			path:   "/home/majo/ws/tools/obsbot-camera-control/CMakeLists.txt",
			loader: LoadCMakeManifest,
			tag:    "@cmake",
			contains: []string{
				"Qt6 direct",
			},
			minDeps: 1,
		},
		{
			name:   "explorer-mcp go.mod",
			path:   "/home/majo/ws/cli/explorer-mcp/go.mod",
			loader: LoadGoModManifest,
			tag:    "@go",
			contains: []string{
				"github.com/joho/godotenv@v1.5.1 direct",
				"github.com/google/jsonschema-go@v0.4.3 indirect",
			},
			minDeps: 5,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := os.Stat(tc.path); err != nil {
				t.Skipf("manifest not available: %v", err)
			}

			root := filepath.Dir(tc.path)
			tag, deps, loaded, err := tc.loader(root, tc.path)
			if err != nil {
				t.Fatalf("loader error: %v", err)
			}
			if !loaded {
				t.Fatal("expected loaded true")
			}
			if tag != tc.tag {
				t.Fatalf("tag = %q, want %q", tag, tc.tag)
			}
			if len(deps) < tc.minDeps {
				t.Fatalf("len(deps) = %d, want >= %d, sample=%v", len(deps), tc.minDeps, deps)
			}

			assertStandardDepScopes(t, deps)

			for _, want := range tc.contains {
				if !slices.Contains(deps, want) {
					t.Fatalf("missing %q in deps (first 8: %v)", want, deps[:min(8, len(deps))])
				}
			}
		})
	}
}

func TestValidateManifestOutputFormat(t *testing.T) {
	t.Parallel()

	validScopes := map[string]struct{}{
		DepScopeDirect:   {},
		DepScopeIndirect: {},
		DepScopeTool:     {},
		DepScopeDev:      {},
	}

	fixtures := map[string]struct {
		loader  func(root, path string) (string, []string, bool, error)
		content string
		want    []string
	}{
		"bun.lock": {
			loader: LoadBunLockManifest,
			content: `{
  "workspaces": { "": { "dependencies": { "zod": "^4.1.12" }, "devDependencies": { "prettier": "3.6.2" } } },
  "packages": { "zod": ["zod@4.1.12", "", {}, "h"], "left-pad": ["left-pad@1.0.0", "", {}, "h"] }
}`,
			want: []string{"zod@^4.1.12 direct", "prettier@3.6.2 dev", "left-pad@1.0.0 indirect"},
		},
		"CMakeLists.txt": {
			loader:  LoadCMakeManifest,
			content: "find_package(Qt6 REQUIRED)\n",
			want:    []string{"Qt6 direct"},
		},
		"deno.json": {
			loader:  LoadDenoManifest,
			content: `{"imports":{"lodash":"npm:lodash@4.17.21"},"devDependencies":{"@std/testing":"jsr:@std/testing@^1.0.0"}}`,
			want:    []string{"lodash@4.17.21 direct", "@std/testing@^1.0.0 dev"},
		},
		"deno.jsonc": {
			loader:  LoadDenoManifest,
			content: `{
  // imports
  "imports": { "lodash": "npm:lodash@4.17.21" },
  "devDependencies": { "@std/testing": "jsr:@std/testing@^1.0.0" }
}`,
			want: []string{"lodash@4.17.21 direct", "@std/testing@^1.0.0 dev"},
		},
		"go.work": {
			loader:  LoadGoWorkManifest,
			content: "go 1.21\n\nuse ./module-a\n",
			want:    []string{"./module-a direct"},
		},
		"composer.json": {
			loader:  LoadComposerManifest,
			content: `{"require":{"php":"^8.1"},"require-dev":{"phpunit/phpunit":"^10.0"}}`,
			want:    []string{"php@^8.1 direct", "phpunit/phpunit@^10.0 dev"},
		},
		"Gemfile": {
			loader:  LoadGemfileManifest,
			content: "gem 'rails', '~> 7.0'\ngroup :development do\n  gem 'rspec'\nend\n",
			want:    []string{"rails@~> 7.0 direct", "rspec dev"},
		},
		"pubspec.yaml": {
			loader:  LoadPubspecManifest,
			content: "dependencies:\n  http: ^1.0.0\ndev_dependencies:\n  flutter_test:\n    sdk: flutter\n",
			want:    []string{"http@^1.0.0 direct", "flutter_test@sdk:flutter dev"},
		},
		"go.mod": {
			loader: LoadGoModManifest,
			content: `module demo

require github.com/foo/bar v1.0.0

tool golang.org/x/tools/cmd/goimports
`,
			want: []string{"github.com/foo/bar@v1.0.0 direct", "golang.org/x/tools/cmd/goimports tool"},
		},
	}

	for name, tc := range fixtures {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			path := filepath.Join(root, name)
			if err := os.WriteFile(path, []byte(tc.content), 0o644); err != nil {
				t.Fatal(err)
			}

			_, deps, loaded, err := tc.loader(root, path)
			if err != nil || !loaded {
				t.Fatalf("loaded=%v err=%v", loaded, err)
			}

			for _, want := range tc.want {
				if !slices.Contains(deps, want) {
					t.Fatalf("missing %q, got %v", want, deps)
				}
			}

			for _, entry := range deps {
				scope := dependencyScope(entry)
				if _, ok := validScopes[scope]; !ok {
					t.Fatalf("invalid scope in %q", entry)
				}
			}
		})
	}
}

func TestGoToolDepsFlagBehavior(t *testing.T) {
	t.Parallel()

	content := `module demo

require golang.org/x/tools v0.30.0

tool golang.org/x/tools/cmd/goimports
`
	withTools := parseGoModDependencies(content, ManifestDepsSettings{ShowGoToolDeps: true})
	withoutTools := parseGoModDependencies(content, ManifestDepsSettings{ShowGoToolDeps: false})

	if !slices.Contains(withTools, "golang.org/x/tools/cmd/goimports@v0.30.0 tool") {
		t.Fatalf("expected tool dep with flag on, got %v", withTools)
	}
	for _, entry := range withoutTools {
		if strings.HasSuffix(entry, " tool") {
			t.Fatalf("tool dep leaked with flag off: %v", withoutTools)
		}
	}
	if !slices.Contains(withoutTools, "golang.org/x/tools@v0.30.0 direct") {
		t.Fatalf("expected require dep preserved, got %v", withoutTools)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
