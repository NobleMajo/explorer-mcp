package deps

import (
	"slices"
	"strings"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestDependenciesFindsManifests(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/go.mod", "module demo\n\nrequire github.com/foo/bar v1.0.0\n")
	testutil.WriteFile(t, root+"/package.json", `{"dependencies":{"left-pad":"1.0.0"}}`)

	testutil.Chdir(t, root)

	result, err := DepsOverview(DefaultSettings())()(root, false)
	if err != nil {
		t.Fatalf("DepsOverview() error: %v", err)
	}

	resp, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	if !slices.Contains(resp, "github.com/foo/bar@v1.0.0 direct") {
		t.Fatalf("missing go dependency, got %v", resp)
	}
	if !slices.Contains(resp, "left-pad@1.0.0 direct") {
		t.Fatalf("missing node dependency, got %v", resp)
	}
}

func TestDependenciesDetectOnlyManifests(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/Cargo.toml", "[package]\nname = \"demo\"\n")
	testutil.WriteFile(t, root+"/pyproject.toml", "[project]\nname = \"demo\"\n")

	testutil.Chdir(t, root)

	result, err := DepsOverview(DefaultSettings())()(root, false)
	if err != nil {
		t.Fatalf("DepsOverview() error: %v", err)
	}

	resp, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if len(resp) != 0 {
		t.Fatalf("expected no parsed dependencies, got %v", resp)
	}
}

func TestDependenciesRequirementsTxt(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/requirements.txt", "flask>=3.0.0\n")
	testutil.Chdir(t, root)

	result, err := DepsOverview(DefaultSettings())()(root, false)
	if err != nil {
		t.Fatalf("DepsOverview() error: %v", err)
	}

	resp, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if len(resp) != 1 || resp[0] != "flask@>=3.0.0 direct" {
		t.Fatalf("unexpected python deps: %v", resp)
	}
}

func TestDependenciesAllManifestLoaders(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/go.mod", `module demo

require (
	github.com/foo/bar v1.0.0
	github.com/indirect/dep v0.1.0 // indirect
)
`)
	testutil.WriteFile(t, root+"/package.json", `{"dependencies":{"left-pad":"1.0.0"},"devDependencies":{"eslint":"9.0.0"}}`)
	testutil.WriteFile(t, root+"/requirements.txt", "flask>=3.0.0\n")
	testutil.WriteFile(t, root+"/Cargo.toml", "[package]\nname = \"demo\"\n")
	testutil.WriteFile(t, root+"/pyproject.toml", "[project]\nname = \"demo\"\n")
	testutil.WriteFile(t, root+"/bun.lock", `{
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
	testutil.WriteFile(t, root+"/CMakeLists.txt", "find_package(Qt6 REQUIRED)\n")
	testutil.WriteFile(t, root+"/deno.json", `{"imports":{"lodash":"npm:lodash@4.17.21"}}`)
	testutil.WriteFile(t, root+"/go.work", "go 1.21\n\nuse ./module-a\n")
	testutil.WriteFile(t, root+"/composer.json", `{"require":{"php":"^8.1"},"require-dev":{"phpunit/phpunit":"^10.0"}}`)
	testutil.WriteFile(t, root+"/Gemfile", "gem 'rails', '~> 7.0'\n")
	testutil.WriteFile(t, root+"/pubspec.yaml", "dependencies:\n  http: ^1.0.0\n")
	testutil.Chdir(t, root)

	result, err := DepsOverview(DefaultSettings())()(root, false)
	if err != nil {
		t.Fatalf("DepsOverview() error: %v", err)
	}

	resp, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	for _, want := range []string{
		"github.com/foo/bar@v1.0.0 direct",
		"github.com/indirect/dep@v0.1.0 indirect",
		"left-pad@1.0.0 direct",
		"eslint@9.0.0 dev",
		"flask@>=3.0.0 direct",
		"zod@^4.1.12 direct",
		"prettier@3.6.2 dev",
		"left-pad@1.0.0 indirect",
		"Qt6 direct",
		"lodash@4.17.21 direct",
		"./module-a direct",
		"php@^8.1 direct",
		"phpunit/phpunit@^10.0 dev",
		"rails@~> 7.0 direct",
		"http@^1.0.0 direct",
	} {
		if !slices.Contains(resp, want) {
			t.Fatalf("missing dependency %q, got %v", want, resp)
		}
	}

	for _, entry := range resp {
		if strings.HasSuffix(entry, " production") || strings.HasSuffix(entry, " development") || strings.Contains(entry, "@direct") || strings.Contains(entry, "@indirect") {
			t.Fatalf("legacy scope label in %q", entry)
		}
	}
}

func TestDependenciesAllScopeTypes(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/go.mod", `module demo

require github.com/indirect/dep v0.1.0 // indirect

tool golang.org/x/tools/cmd/goimports
`)
	testutil.WriteFile(t, root+"/package.json", `{"devDependencies":{"eslint":"9.0.0"}}`)
	testutil.WriteFile(t, root+"/bun.lock", `{
  "workspaces": { "": { "dependencies": { "zod": "^4.1.12" } } },
  "packages": {
    "zod": ["zod@4.1.12", "", {}, "hash"],
    "left-pad": ["left-pad@1.0.0", "", {}, "hash"]
  }
}`)
	testutil.Chdir(t, root)

	result, err := DepsOverview(DefaultSettings())()(root, false)
	if err != nil {
		t.Fatalf("DepsOverview() error: %v", err)
	}

	resp, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	for _, want := range []string{
		"github.com/indirect/dep@v0.1.0 indirect",
		"golang.org/x/tools/cmd/goimports tool",
		"eslint@9.0.0 dev",
		"zod@^4.1.12 direct",
		"left-pad@1.0.0 indirect",
	} {
		if !slices.Contains(resp, want) {
			t.Fatalf("missing dependency %q, got %v", want, resp)
		}
	}
}

func TestDependenciesSkipsMissingManifests(t *testing.T) {
	root := t.TempDir()
	testutil.Chdir(t, root)

	result, err := DepsOverview(DefaultSettings())()(root, false)
	if err != nil {
		t.Fatalf("DepsOverview() error: %v", err)
	}

	resp, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if len(resp) != 0 {
		t.Fatalf("expected no dependencies, got %v", resp)
	}
}

func TestDependenciesGoModLoaderError(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/go.mod", "")
	testutil.Chdir(t, root)

	result, err := DepsOverview(DefaultSettings())()(root, false)
	if err != nil {
		t.Fatalf("DepsOverview() error: %v", err)
	}

	resp, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if len(resp) != 0 {
		t.Fatalf("expected empty go.mod to yield no dependencies, got %v", resp)
	}
}

func TestDependenciesGoToolDeps(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/go.mod", `module demo

require (
	github.com/foo/bar v1.0.0
	golang.org/x/tools v0.30.0
)

tool (
	golang.org/x/tools/cmd/goimports
)
`)
	testutil.Chdir(t, root)

	result, err := DepsOverview(Settings{ShowGoToolDeps: true})()(root, false)
	if err != nil {
		t.Fatalf("DepsOverview() error: %v", err)
	}

	resp, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	for _, want := range []string{
		"github.com/foo/bar@v1.0.0 direct",
		"golang.org/x/tools@v0.30.0 direct",
		"golang.org/x/tools/cmd/goimports@v0.30.0 tool",
	} {
		if !slices.Contains(resp, want) {
			t.Fatalf("missing dependency %q, got %v", want, resp)
		}
	}
}

func TestDependenciesGoToolDepsHiddenWhenDisabled(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/go.mod", `module demo

require golang.org/x/tools v0.30.0

tool golang.org/x/tools/cmd/goimports
`)
	testutil.Chdir(t, root)

	result, err := DepsOverview(Settings{ShowGoToolDeps: false})()(root, false)
	if err != nil {
		t.Fatalf("DepsOverview() error: %v", err)
	}

	resp, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if !slices.Contains(resp, "golang.org/x/tools@v0.30.0 direct") {
		t.Fatalf("missing require dependency, got %v", resp)
	}
	for _, entry := range resp {
		if strings.HasSuffix(entry, " tool") {
			t.Fatalf("expected tool deps hidden, got %v", resp)
		}
	}
}

func TestDependenciesLoaderError(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/package.json", `{invalid`)

	testutil.Chdir(t, root)

	_, err := DepsOverview(DefaultSettings())()(root, false)
	if err == nil {
		t.Fatal("expected error for invalid package.json")
	}
}

func TestDefaultSettingsMatchesGlobalsDefaults(t *testing.T) {
	settings := DefaultSettings()
	if !settings.ShowGoToolDeps {
		t.Fatal("expected ShowGoToolDeps true in deps.DefaultSettings")
	}
	if settings.ShowGoToolDeps != globals.DefaultManifestDepsSettings().ShowGoToolDeps {
		t.Fatal("deps settings out of sync with globals defaults")
	}
}

func TestDependenciesPackageJsonDevScope(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/package.json", `{"devDependencies":{"eslint":"9.0.0"}}`)
	testutil.Chdir(t, root)

	result, err := DepsOverview(DefaultSettings())()(root, false)
	if err != nil {
		t.Fatalf("DepsOverview() error: %v", err)
	}

	resp, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if len(resp) != 1 || resp[0] != "eslint@9.0.0 dev" {
		t.Fatalf("unexpected deps: %v", resp)
	}
}
