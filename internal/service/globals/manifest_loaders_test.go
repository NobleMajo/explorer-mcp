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
		"CMakeLists.txt",
		"Cargo.toml",
		"Gemfile",
		"bun.lock",
		"composer.json",
		"deno.json",
		"deno.jsonc",
		"go.mod",
		"go.work",
		"package.json",
		"pubspec.yaml",
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
		"go.mod": `module demo

require github.com/foo/bar v1.0.0

tool golang.org/x/tools/cmd/goimports
`,
		"package.json":     `{"dependencies":{"left-pad":"1.0.0"}}`,
		"requirements.txt": "flask>=3.0.0\n",
		"Cargo.toml":       "[package]\nname = \"demo\"\n",
		"pyproject.toml":   "[project]\nname = \"demo\"\n",
		"bun.lock": `{
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
}`,
		"CMakeLists.txt": `cmake_minimum_required(VERSION 3.16)
find_package(Qt6 REQUIRED COMPONENTS Widgets)
FetchContent_Declare(fmt)
`,
		"deno.json": `{
  "imports": { "lodash": "npm:lodash@4.17.21" },
  "dependencies": { "@std/path": "jsr:@std/path@^1.0.0" },
  "devDependencies": { "@std/testing": "jsr:@std/testing@^1.0.0" }
}`,
		"deno.jsonc": `{
  // config
  "imports": { "lodash": "npm:lodash@4.17.21" }
}`,
		"go.work": `go 1.21

use (
	./module-a
	./module-b
)
`,
		"composer.json": `{
  "require": { "laravel/framework": "^10.0" },
  "require-dev": { "phpunit/phpunit": "^10.0" }
}`,
		"Gemfile": `source 'https://rubygems.org'

gem 'rails', '~> 7.0'
group :development do
  gem 'rspec', '~> 3.0'
end
`,
		"pubspec.yaml": `name: demo

dependencies:
  http: ^1.0.0
  flutter:
    sdk: flutter

dev_dependencies:
  flutter_test:
    sdk: flutter
`,
	}

	expectedTags := map[string]string{
		"go.mod":           "@go",
		"package.json":     "@npm",
		"requirements.txt": "@pip",
		"Cargo.toml":       "@cargo",
		"pyproject.toml":   "@python",
		"bun.lock":         "@bun",
		"CMakeLists.txt":   "@cmake",
		"deno.json":        "@deno",
		"deno.jsonc":       "@deno",
		"go.work":          "@go-workspace",
		"composer.json":    "@composer",
		"Gemfile":          "@ruby",
		"pubspec.yaml":     "@dart",
	}

	expectedContains := map[string][]string{
		"go.mod":           {"github.com/foo/bar@v1.0.0 direct", "golang.org/x/tools/cmd/goimports tool"},
		"package.json":     {"left-pad@1.0.0 direct"},
		"requirements.txt": {"flask@>=3.0.0 direct"},
		"bun.lock": {
			"zod@^4.1.12 direct",
			"prettier@3.6.2 dev",
			"left-pad@1.0.0 indirect",
		},
		"CMakeLists.txt": {"Qt6 direct", "fmt direct"},
		"deno.json": {
			"lodash@4.17.21 direct",
			"@std/path@^1.0.0 direct",
			"@std/testing@^1.0.0 dev",
		},
		"deno.jsonc":     {"lodash@4.17.21 direct"},
		"go.work":        {"./module-a direct", "./module-b direct"},
		"composer.json":  {"laravel/framework@^10.0 direct", "phpunit/phpunit@^10.0 dev"},
		"Gemfile":        {"rails@~> 7.0 direct", "rspec@~> 3.0 dev"},
		"pubspec.yaml":   {"http@^1.0.0 direct", "flutter@sdk:flutter direct", "flutter_test@sdk:flutter dev"},
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

		want, ok := expectedContains[fileName]
		if !ok {
			if len(got) != 0 {
				t.Fatalf("%q dependencies = %v, want empty", fileName, got)
			}
			continue
		}
		for _, entry := range want {
			if !slices.Contains(got, entry) {
				t.Fatalf("%q missing dependency %q in %v", fileName, entry, got)
			}
		}
	}
}
