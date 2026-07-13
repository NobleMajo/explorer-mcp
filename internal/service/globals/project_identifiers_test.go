package globals

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCollectProjectIdentifierFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		subfiles []string
		want     []string
	}{
		{
			name:     "makefile",
			subfiles: []string{"Makefile", "README.md"},
			want:     []string{"@makefile"},
		},
		{
			name:     "gnumakefile",
			subfiles: []string{"GNUmakefile"},
			want:     []string{"@makefile"},
		},
		{
			name:     "makefile lowercase",
			subfiles: []string{"makefile"},
			want:     []string{"@makefile"},
		},
		{
			name:     "tsconfig json",
			subfiles: []string{"tsconfig.json"},
			want:     []string{"@tsconfig"},
		},
		{
			name:     "prefixed tsconfig",
			subfiles: []string{"app.tsconfig.json"},
			want:     []string{"@tsconfig"},
		},
		{
			name:     "tsconfig env file",
			subfiles: []string{"tsconfig.app.json"},
			want:     []string{"@tsconfig"},
		},
		{
			name:     "angular",
			subfiles: []string{"angular.json"},
			want:     []string{"@angular"},
		},
		{
			name:     "dockerfile",
			subfiles: []string{"Dockerfile"},
			want:     []string{"@docker"},
		},
		{
			name:     "dockerfile variant",
			subfiles: []string{"Dockerfile.dev"},
			want:     []string{"@docker"},
		},
		{
			name:     "docker compose yaml",
			subfiles: []string{"docker-compose.yaml"},
			want:     []string{"@docker-compose"},
		},
		{
			name:     "podman compose",
			subfiles: []string{"podman-compose.yml"},
			want:     []string{"@docker-compose"},
		},
		{
			name:     "containerfile variant",
			subfiles: []string{"Containerfile.dev"},
			want:     []string{"@docker"},
		},
		{
			name:     "docker compose override",
			subfiles: []string{"docker-compose-dev.yml"},
			want:     []string{"@docker-compose"},
		},
		{
			name:     "vite",
			subfiles: []string{"vite.config.ts"},
			want:     []string{"@vite"},
		},
		{
			name:     "biome",
			subfiles: []string{"biome.json"},
			want:     []string{"@biome"},
		},
		{
			name:     "biome jsonc",
			subfiles: []string{"biome.jsonc"},
			want:     []string{"@biome"},
		},
		{
			name:     "taskfile dist",
			subfiles: []string{"Taskfile.dist.yml"},
			want:     []string{"@task"},
		},
		{
			name:     "taskfile",
			subfiles: []string{"Taskfile.yml"},
			want:     []string{"@task"},
		},
		{
			name:     "next",
			subfiles: []string{"next.config.mjs"},
			want:     []string{"@next"},
		},
		{
			name:     "nuxt",
			subfiles: []string{"nuxt.config.ts"},
			want:     []string{"@nuxt"},
		},
		{
			name:     "svelte",
			subfiles: []string{"svelte.config.js"},
			want:     []string{"@svelte"},
		},
		{
			name:     "turbo",
			subfiles: []string{"turbo.json"},
			want:     []string{"@turbo"},
		},
		{
			name:     "nx",
			subfiles: []string{"nx.json"},
			want:     []string{"@nx"},
		},
		{
			name:     "pnpm workspace",
			subfiles: []string{"pnpm-workspace.yaml"},
			want:     []string{"@pnpm-workspace"},
		},
		{
			name:     "just",
			subfiles: []string{"justfile"},
			want:     []string{"@just"},
		},
		{
			name:     "meson",
			subfiles: []string{"meson.build"},
			want:     []string{"@meson"},
		},
		{
			name:     "gradle",
			subfiles: []string{"build.gradle.kts"},
			want:     []string{"@gradle"},
		},
		{
			name:     "maven",
			subfiles: []string{"pom.xml"},
			want:     []string{"@maven"},
		},
		{
			name:     "nix",
			subfiles: []string{"flake.nix"},
			want:     []string{"@nix"},
		},
		{
			name:     "bazel",
			subfiles: []string{"MODULE.bazel"},
			want:     []string{"@bazel"},
		},
		{
			name:     "tauri",
			subfiles: []string{"tauri.conf.json"},
			want:     []string{"@tauri"},
		},
		{
			name:     "cloudflare",
			subfiles: []string{"wrangler.toml"},
			want:     []string{"@cloudflare"},
		},
		{
			name:     "rake",
			subfiles: []string{"Rakefile"},
			want:     []string{"@rake"},
		},
		{
			name:     "earthly",
			subfiles: []string{"Earthfile"},
			want:     []string{"@earthly"},
		},
		{
			name:     "combined",
			subfiles: []string{"Makefile", "angular.json", "tsconfig.json", "Dockerfile", "docker-compose.yml", "vite.config.ts"},
			want:     []string{"@makefile", "@tsconfig", "@angular", "@docker", "@docker-compose", "@vite"},
		},
		{
			name:     "none",
			subfiles: []string{"README.md", "next.config.backup", "vite.config.backup"},
			want:     nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := CollectProjectIdentifierFlags("/tmp/project", tc.subfiles, nil)
			if err != nil {
				t.Fatalf("CollectProjectIdentifierFlags() error: %v", err)
			}
			if len(tc.want) == 0 {
				if len(got) != 0 {
					t.Fatalf("CollectProjectIdentifierFlags() = %#v, want empty", got)
				}
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("CollectProjectIdentifierFlags() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestIsTSConfigFileName(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"tsconfig.json", "app.tsconfig.json", "tsconfig.app.json"} {
		if !isTSConfigFileName(name) {
			t.Fatalf("isTSConfigFileName(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"package.json", "config.json", "nontsconfig.app.json"} {
		if isTSConfigFileName(name) {
			t.Fatalf("isTSConfigFileName(%q) = true, want false", name)
		}
	}
}

func TestIsDockerComposeFileName(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"compose.yml",
		"docker-compose.yaml",
		"docker-compose-dev.yml",
		"docker-compose.playwright-mcp.yml",
		"compose.override.yaml",
		"compose-dev.yml",
		"podman-compose.yml",
	} {
		if !isDockerComposeFileName(name) {
			t.Fatalf("isDockerComposeFileName(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"docker-compose", "compose.txt", "my-compose.yml"} {
		if isDockerComposeFileName(name) {
			t.Fatalf("isDockerComposeFileName(%q) = true, want false", name)
		}
	}
}

func TestIsDockerfileName(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"Dockerfile", "Containerfile", "Dockerfile.dev", "Containerfile.dev", "app.dockerfile"} {
		if !isDockerfileName(name) {
			t.Fatalf("isDockerfileName(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"dockerfile", "Makefile", "compose.yml"} {
		if isDockerfileName(name) {
			t.Fatalf("isDockerfileName(%q) = true, want false", name)
		}
	}
}

func TestHasNamedConfigFile(t *testing.T) {
	t.Parallel()

	subfiles := []string{"vite.config.ts", "next.config.mjs", "README.md"}
	if !hasNamedConfigFile(subfiles, "vite.config") {
		t.Fatal("expected vite.config match")
	}
	if !hasNamedConfigFile(subfiles, "next.config") {
		t.Fatal("expected next.config match")
	}
	if hasNamedConfigFile(subfiles, "rollup.config") {
		t.Fatal("expected no rollup.config match")
	}
	if isNamedConfigFile("next.config.backup", "next.config") {
		t.Fatal("expected next.config.backup to be rejected")
	}
	if !isNamedConfigFile("vite.config.mts", "vite.config") {
		t.Fatal("expected vite.config.mts match")
	}
}

func TestProjectIdentifiersRegistry(t *testing.T) {
	t.Parallel()

	if len(ProjectIdentifiers) != 24 {
		t.Fatalf("len(ProjectIdentifiers) = %d, want 24", len(ProjectIdentifiers))
	}

	registry := []struct {
		tag      string
		subfiles []string
	}{
		{tag: "@makefile", subfiles: []string{"GNUmakefile"}},
		{tag: "@tsconfig", subfiles: []string{"tsconfig.json"}},
		{tag: "@angular", subfiles: []string{"angular.json"}},
		{tag: "@docker", subfiles: []string{"Dockerfile"}},
		{tag: "@docker-compose", subfiles: []string{"docker-compose.yaml"}},
		{tag: "@vite", subfiles: []string{"vite.config.ts"}},
		{tag: "@biome", subfiles: []string{"biome.jsonc"}},
		{tag: "@task", subfiles: []string{"Taskfile.dist.yml"}},
		{tag: "@next", subfiles: []string{"next.config.mjs"}},
		{tag: "@nuxt", subfiles: []string{"nuxt.config.ts"}},
		{tag: "@svelte", subfiles: []string{"svelte.config.js"}},
		{tag: "@turbo", subfiles: []string{"turbo.json"}},
		{tag: "@nx", subfiles: []string{"nx.json"}},
		{tag: "@pnpm-workspace", subfiles: []string{"pnpm-workspace.yaml"}},
		{tag: "@just", subfiles: []string{"justfile"}},
		{tag: "@meson", subfiles: []string{"meson.build"}},
		{tag: "@gradle", subfiles: []string{"settings.gradle.kts"}},
		{tag: "@maven", subfiles: []string{"pom.xml"}},
		{tag: "@nix", subfiles: []string{"flake.nix"}},
		{tag: "@bazel", subfiles: []string{"MODULE.bazel"}},
		{tag: "@tauri", subfiles: []string{"tauri.conf.json"}},
		{tag: "@cloudflare", subfiles: []string{"wrangler.toml"}},
		{tag: "@rake", subfiles: []string{"Rakefile"}},
		{tag: "@earthly", subfiles: []string{"Earthfile"}},
	}

	if len(registry) != len(ProjectIdentifiers) {
		t.Fatalf("registry cases = %d, want %d", len(registry), len(ProjectIdentifiers))
	}

	for _, tc := range registry {
		tc := tc
		t.Run(tc.tag, func(t *testing.T) {
			t.Parallel()
			got, err := CollectProjectIdentifierFlags("/tmp/project", tc.subfiles, nil)
			if err != nil {
				t.Fatalf("CollectProjectIdentifierFlags() error: %v", err)
			}
			if len(got) != 1 || got[0] != tc.tag {
				t.Fatalf("CollectProjectIdentifierFlags() = %#v, want [%q]", got, tc.tag)
			}
		})
	}
}

func TestIdentifyByFileNames(t *testing.T) {
	t.Parallel()

	identify := identifyByFileNames("@demo", "alpha.txt", "beta.txt")
	got, err := identify("/tmp", []string{"beta.txt"}, nil)
	if err != nil {
		t.Fatalf("identifyByFileNames() error: %v", err)
	}
	if len(got) != 1 || got[0] != "@demo" {
		t.Fatalf("identifyByFileNames() = %#v, want [@demo]", got)
	}

	got, err = identify("/tmp", []string{"gamma.txt"}, nil)
	if err != nil {
		t.Fatalf("identifyByFileNames() error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("identifyByFileNames() = %#v, want empty", got)
	}
}

func TestIdentifyByMatcher(t *testing.T) {
	t.Parallel()

	identify := identifyByMatcher("@demo", func(name string) bool {
		return strings.HasSuffix(name, ".demo")
	})
	got, err := identify("/tmp", []string{"app.demo"}, nil)
	if err != nil {
		t.Fatalf("identifyByMatcher() error: %v", err)
	}
	if len(got) != 1 || got[0] != "@demo" {
		t.Fatalf("identifyByMatcher() = %#v, want [@demo]", got)
	}
}

func TestIsGradleFileName(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"build.gradle", "build.gradle.kts", "settings.gradle", "settings.gradle.kts"} {
		if !isGradleFileName(name) {
			t.Fatalf("isGradleFileName(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"gradle.properties", "gradlew", "build.xml"} {
		if isGradleFileName(name) {
			t.Fatalf("isGradleFileName(%q) = true, want false", name)
		}
	}
}

func TestCollectProjectIdentifierFlagsRealWSPaths(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("home dir unavailable")
	}
	ws := filepath.Join(home, "ws")
	if _, err := os.Stat(ws); err != nil {
		t.Skip("~/ws not available")
	}

	tests := []struct {
		path string
		want []string
	}{
		{filepath.Join(ws, "cli", "explorer-mcp"), []string{"@makefile", "@docker", "@docker-compose"}},
		{filepath.Join(ws, "cli", "deployit"), []string{"@makefile", "@docker", "@docker-compose"}},
		{filepath.Join(ws, "web", "twitchbot2"), []string{"@makefile", "@docker", "@docker-compose", "@tsconfig"}},
		{filepath.Join(ws, "web", "mobidesk", "ui"), []string{"@vite", "@biome", "@tsconfig"}},
		{filepath.Join(ws, "web", "nobletool", "ui"), []string{"@vite", "@tsconfig"}},
		{filepath.Join(ws, "web", "angular-test-app"), []string{"@angular", "@tsconfig"}},
		{filepath.Join(ws, "tests", "golang-webui", "golang-webui"), []string{"@task"}},
		{filepath.Join(ws, "services", "door"), []string{"@makefile", "@docker-compose"}},
		{filepath.Join(ws, "games", "verseforge", "config"), []string{"@docker", "@docker-compose"}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(filepath.Base(tc.path), func(t *testing.T) {
			entries, err := os.ReadDir(tc.path)
			if err != nil {
				t.Skipf("%s not available: %v", tc.path, err)
			}

			subfiles := make([]string, 0, len(entries))
			subdirs := make([]string, 0)
			for _, entry := range entries {
				if entry.IsDir() {
					subdirs = append(subdirs, entry.Name())
					continue
				}
				subfiles = append(subfiles, entry.Name())
			}

			got, err := CollectProjectIdentifierFlags(tc.path, subfiles, subdirs)
			if err != nil {
				t.Fatalf("CollectProjectIdentifierFlags() error: %v", err)
			}
			for _, tag := range tc.want {
				found := false
				for _, g := range got {
					if g == tag {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("CollectProjectIdentifierFlags(%q) = %#v, missing %q", tc.path, got, tag)
				}
			}
		})
	}
}
