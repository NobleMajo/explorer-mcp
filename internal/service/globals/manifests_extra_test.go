package globals

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestParseGoWorkUse(t *testing.T) {
	t.Parallel()

	content := `go 1.21

use (
	./module-a
	./module-b
)

use ./single
`
	got := parseGoWorkUse(content)
	want := []string{"./module-a direct", "./module-b direct", "./single direct"}
	if len(got) != len(want) {
		t.Fatalf("parseGoWorkUse() = %v, want %v", got, want)
	}
	for _, entry := range want {
		if !slices.Contains(got, entry) {
			t.Fatalf("missing %q in %v", entry, got)
		}
	}
}

func TestParseGemfile(t *testing.T) {
	t.Parallel()

	content := `gem 'rails', '~> 7.0'
group :development do
  gem "rspec", "~> 3.0"
end
`
	got := parseGemfile(content)
	want := []string{"rails@~> 7.0 direct", "rspec@~> 3.0 dev"}
	for _, entry := range want {
		if !slices.Contains(got, entry) {
			t.Fatalf("missing %q in %v", entry, got)
		}
	}
}

func TestParsePubspecDependencies(t *testing.T) {
	t.Parallel()

	content := `dependencies:
  http: ^1.0.0

dev_dependencies:
  flutter_test:
    sdk: flutter
`
	got := parsePubspecDependencies(content)
	want := []string{"http@^1.0.0 direct", "flutter_test@sdk:flutter dev"}
	for _, entry := range want {
		if !slices.Contains(got, entry) {
			t.Fatalf("missing %q in %v", entry, got)
		}
	}
}

func TestLoadBunLockManifestTrailingComma(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "bun.lock")
	testutil.WriteFile(t, path, `{
  "workspaces": {
    "": {
      "dependencies": { "zod": "^4.1.12", },
    }
  },
  "packages": {
    "zod": ["zod@4.1.12", "", {}, "hash"],
  }
}`)

	tag, got, loaded, err := LoadBunLockManifest(root, path)
	if err != nil {
		t.Fatalf("LoadBunLockManifest() error: %v", err)
	}
	if !loaded {
		t.Fatal("expected loaded true")
	}
	if tag != "@bun" {
		t.Fatalf("tag = %q, want @bun", tag)
	}
	if !slices.Contains(got, "zod@^4.1.12 direct") {
		t.Fatalf("missing direct zod dependency, got %v", got)
	}
}

func TestLoadCMakeManifest(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "CMakeLists.txt")
	testutil.WriteFile(t, path, "find_package(Qt6 REQUIRED)\nFetchContent_Declare(fmt)\n")

	tag, got, loaded, err := LoadCMakeManifest(root, path)
	if err != nil {
		t.Fatalf("LoadCMakeManifest() error: %v", err)
	}
	if !loaded || tag != "@cmake" {
		t.Fatalf("loaded=%v tag=%q, want loaded true tag @cmake", loaded, tag)
	}
	for _, want := range []string{"Qt6 direct", "fmt direct"} {
		if !slices.Contains(got, want) {
			t.Fatalf("missing %q in %v", want, got)
		}
	}
}

func TestLoadDenoManifestJSONC(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "deno.jsonc")
	testutil.WriteFile(t, path, `{
  // imports
  "imports": { "lodash": "npm:lodash@4.17.21" }
}`)

	tag, got, loaded, err := LoadDenoManifest(root, path)
	if err != nil {
		t.Fatalf("LoadDenoManifest() error: %v", err)
	}
	if !loaded || tag != "@deno" {
		t.Fatalf("loaded=%v tag=%q, want loaded true tag @deno", loaded, tag)
	}
	if !slices.Contains(got, "lodash@4.17.21 direct") {
		t.Fatalf("unexpected result: %v", got)
	}
}

func TestLoadComposerManifest(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "composer.json")
	testutil.WriteFile(t, path, `{"require":{"php":"^8.1"},"require-dev":{"phpunit/phpunit":"^10.0"}}`)

	tag, got, loaded, err := LoadComposerManifest(root, path)
	if err != nil {
		t.Fatalf("LoadComposerManifest() error: %v", err)
	}
	if !loaded || tag != "@composer" {
		t.Fatalf("loaded=%v tag=%q, want loaded true tag @composer", loaded, tag)
	}
	for _, want := range []string{"php@^8.1 direct", "phpunit/phpunit@^10.0 dev"} {
		if !slices.Contains(got, want) {
			t.Fatalf("missing %q in %v", want, got)
		}
	}
}

func TestLoadGoWorkManifest(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "go.work")
	testutil.WriteFile(t, path, "go 1.21\n\nuse ./demo\n")

	tag, got, loaded, err := LoadGoWorkManifest(root, path)
	if err != nil {
		t.Fatalf("LoadGoWorkManifest() error: %v", err)
	}
	if !loaded || tag != "@go-workspace" {
		t.Fatalf("loaded=%v tag=%q, want loaded true tag @go-workspace", loaded, tag)
	}
	if len(got) != 1 || got[0] != "./demo direct" {
		t.Fatalf("unexpected result: %v", got)
	}
}

func TestLoadGemfileManifest(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "Gemfile")
	testutil.WriteFile(t, path, "gem 'sinatra', '~> 3.0'\n")

	tag, got, loaded, err := LoadGemfileManifest(root, path)
	if err != nil {
		t.Fatalf("LoadGemfileManifest() error: %v", err)
	}
	if !loaded || tag != "@ruby" {
		t.Fatalf("loaded=%v tag=%q, want loaded true tag @ruby", loaded, tag)
	}
	if len(got) != 1 || got[0] != "sinatra@~> 3.0 direct" {
		t.Fatalf("unexpected result: %v", got)
	}
}

func TestLoadPubspecManifest(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "pubspec.yaml")
	testutil.WriteFile(t, path, "dependencies:\n  http: ^1.0.0\n")

	tag, got, loaded, err := LoadPubspecManifest(root, path)
	if err != nil {
		t.Fatalf("LoadPubspecManifest() error: %v", err)
	}
	if !loaded || tag != "@dart" {
		t.Fatalf("loaded=%v tag=%q, want loaded true tag @dart", loaded, tag)
	}
	if len(got) != 1 || got[0] != "http@^1.0.0 direct" {
		t.Fatalf("unexpected result: %v", got)
	}
}

func TestLoadBunLockManifestMissingFile(t *testing.T) {
	t.Parallel()

	_, _, loaded, err := LoadBunLockManifest(t.TempDir(), "/does/not/exist/bun.lock")
	if err != nil {
		t.Fatalf("LoadBunLockManifest() error: %v", err)
	}
	if loaded {
		t.Fatal("expected loaded false for missing bun.lock")
	}
}
