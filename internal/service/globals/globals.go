package globals

import (
	"path/filepath"
	"slices"
	"sort"
)

type ManifestDepsSettings struct {
	ShowGoToolDeps bool
}

func DefaultManifestDepsSettings() ManifestDepsSettings {
	return ManifestDepsSettings{ShowGoToolDeps: true}
}

type ManifestLoader func(root, manifestPath string) (string, []string, bool, error)

type ProjectIdentifier func(path string, subfiles []string, subdirs []string) ([]string, error)

var KnownContainerFileNames = []string{
	"compose.yml",
	"compose.yaml",
	"docker-compose.yml",
	"docker-compose.yaml",
	"Dockerfile",
	"Containerfile",
	"Podmanfile",
	"Earthfile",
	"Kube-compose.yml",
	"vagranteverywhere.yml",
	"podman-compose.yml",
	"skaffold.yaml",
	"Tiltfile",
	"devcontainer.json",
	"*.dockerfile",
	"*.Dockerfile",
	"Dockerfile.*",
	"Containerfile.*",
	"docker-compose.*.yml",
	"docker-compose.*.yaml",
	"compose.*.yml",
	"compose.*.yaml",
}

var KnownContainerDirectoryNames = []string{
	"container",
	".container",
	"docker",
	".docker",
	"podman",
	".podman",
	".devcontainer",
	"deploy",
}

var KnownContainerCLINames = []string{
	"docker",
	"podman",
	"nerdctl",
	"buildah",
	"skopeo",
	"docker-compose",
	"podman-compose",
	"compose",
}

var KnownContainerRuntimeCLINames = []string{
	"docker",
	"podman",
	"nerdctl",
}

var CommonCLIToolNames = []string{
	"npm",
	"yarn",
	"pnpm",
	"bun",
	"pip",
	"pip3",
	"poetry",
	"uv",
	"cargo",
	"go",
	"gem",
	"composer",
	"node",
	"python",
	"python3",
	"ruby",
	"java",
	"dotnet",
	"deno",
	"gcc",
	"g++",
	"clang",
	"rustc",
	"javac",
	"tsc",
	"make",
	"cmake",
	"ninja",
	"mage",
	"gradle",
	"mvn",
	"bazel",
	"bash",
	"sh",
	"zsh",
	"fish",
	"pwsh",
	"perl",
	"git",
	"gh",
	"docker",
	"docker-compose",
	"kubectl",
	"helm",
}

var ManifestLoaders = map[string]ManifestLoader{
	"CMakeLists.txt":   LoadCMakeManifest,
	"Cargo.toml":       LoadCargoManifest,
	"Gemfile":          LoadGemfileManifest,
	"bun.lock":         LoadBunLockManifest,
	"composer.json":    LoadComposerManifest,
	"deno.json":        LoadDenoManifest,
	"deno.jsonc":       LoadDenoManifest,
	"go.mod":           LoadGoModManifest,
	"go.work":          LoadGoWorkManifest,
	"package.json":     LoadPackageJsonManifest,
	"pubspec.yaml":     LoadPubspecManifest,
	"pyproject.toml":   LoadPyprojectManifest,
	"requirements.txt": LoadRequirementsManifest,
}

var ProjectIdentifiers = []ProjectIdentifier{
	identifyMakefileProject,
	identifyByMatcher("@tsconfig", isTSConfigFileName),
	identifyByFileNames("@angular", "angular.json"),
	identifyByMatcher("@docker", isDockerfileName),
	identifyByMatcher("@docker-compose", isDockerComposeFileName),
	identifyByMatcher("@vite", func(name string) bool { return isNamedConfigFile(name, "vite.config") }),
	identifyByFileNames("@biome", "biome.json", "biome.jsonc"),
	identifyByFileNames("@task", "Taskfile.yml", "Taskfile.yaml", "Taskfile.dist.yml", "Taskfile.dist.yaml"),
	identifyByMatcher("@next", func(name string) bool { return isNamedConfigFile(name, "next.config") }),
	identifyByMatcher("@nuxt", func(name string) bool { return isNamedConfigFile(name, "nuxt.config") }),
	identifyByMatcher("@svelte", func(name string) bool { return isNamedConfigFile(name, "svelte.config") }),
	identifyByFileNames("@turbo", "turbo.json"),
	identifyByFileNames("@nx", "nx.json"),
	identifyByFileNames("@pnpm-workspace", "pnpm-workspace.yaml"),
	identifyByFileNames("@just", "justfile", "Justfile"),
	identifyByFileNames("@meson", "meson.build"),
	identifyByMatcher("@gradle", isGradleFileName),
	identifyByFileNames("@maven", "pom.xml"),
	identifyByFileNames("@nix", "flake.nix", "shell.nix", "default.nix"),
	identifyByFileNames("@bazel", "WORKSPACE", "MODULE.bazel", "BUILD.bazel"),
	identifyByFileNames("@tauri", "tauri.conf.json"),
	identifyByFileNames("@cloudflare", "wrangler.toml"),
	identifyByFileNames("@rake", "Rakefile"),
	identifyByFileNames("@earthly", "Earthfile"),
}

var ScanIgnoreFiles = []string{
	".tmp",
	"tmp",
	".git",
	".angular",
	".cache",
	"cache",
	"__pycache__",
	".pytest_cache",
	".mypy_cache",
	".ruff_cache",
	".tox",
	".nox",
	".turbo",
	".sass-cache",
	".gradle",
	"htmlcov",
	".eslintcache",
	".stylelintcache",
	".tsbuildinfo",
}

func ManifestLoaderFileNames() []string {
	fileNames := make([]string, 0, len(ManifestLoaders))
	for fileName := range ManifestLoaders {
		fileNames = append(fileNames, fileName)
	}
	sort.Strings(fileNames)
	return fileNames
}

func CollectManifestFlags(path string, subfiles []string) ([]string, error) {
	flags := make([]string, 0)
	for _, manifestFileName := range ManifestLoaderFileNames() {
		if !slices.Contains(subfiles, manifestFileName) {
			continue
		}
		tag, _, loaded, err := ManifestLoaders[manifestFileName](path, filepath.Join(path, manifestFileName))
		if err != nil {
			return nil, err
		}
		if loaded && tag != "" {
			flags = append(flags, tag)
		}
	}
	return flags, nil
}

func CollectManifestDependencies(root string, settings ManifestDepsSettings) ([]string, error) {
	dependencies := make([]string, 0)
	for _, fileName := range ManifestLoaderFileNames() {
		manifestPath := filepath.Join(root, fileName)
		var entries []string
		var loaded bool
		var err error

		if fileName == "go.mod" {
			_, entries, loaded, err = loadGoModManifest(root, manifestPath, settings)
		} else {
			_, entries, loaded, err = ManifestLoaders[fileName](root, manifestPath)
		}
		if err != nil {
			return nil, err
		}
		if !loaded {
			continue
		}
		dependencies = append(dependencies, entries...)
	}
	sort.Strings(dependencies)
	return dependencies, nil
}

func CollectSiblingProjectFlags(path string, subfiles, subdirs []string) ([]string, error) {
	flags := make([]string, 0)
	if slices.Contains(subdirs, ".git") {
		flags = append(flags, "@git")
	}

	manifestFlags, err := CollectManifestFlags(path, subfiles)
	if err != nil {
		return nil, err
	}
	flags = append(flags, manifestFlags...)

	identifierFlags, err := CollectProjectIdentifierFlags(path, subfiles, subdirs)
	if err != nil {
		return nil, err
	}
	flags = append(flags, identifierFlags...)

	return flags, nil
}

func CollectProjectIdentifierFlags(path string, subfiles []string, subdirs []string) ([]string, error) {
	flags := make([]string, 0)
	for _, identify := range ProjectIdentifiers {
		found, err := identify(path, subfiles, subdirs)
		if err != nil {
			return nil, err
		}
		flags = append(flags, found...)
	}
	return flags, nil
}

func IsScanIgnored(name string) bool {
	for _, ignored := range ScanIgnoreFiles {
		if ignored == name {
			return true
		}
	}
	return false
}
