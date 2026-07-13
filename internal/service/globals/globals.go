package globals

import (
	"os"
	"path/filepath"
	"sort"
)

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
	"docker",
	".devcontainer",
	"deploy",
	"containers",
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

type ManifestLoader func(root, manifestPath string) ([]string, error)

var ManifestLoaders = map[string]ManifestLoader{
	"go.mod":           LoadGoModManifest,
	"package.json":     LoadPackageJsonManifest,
	"requirements.txt": LoadRequirementsManifest,
	"Cargo.toml":       LoadCargoManifest,
	"pyproject.toml":   LoadPyprojectManifest,
}

var ManifestLoaderTags = map[string]string{
	"go.mod":           "@go",
	"package.json":     "@npm",
	"requirements.txt": "@pip",
	"Cargo.toml":       "@cargo",
	"pyproject.toml":   "@python",
}

func ManifestLoaderFileNames() []string {
	fileNames := make([]string, 0, len(ManifestLoaders))
	for fileName := range ManifestLoaders {
		fileNames = append(fileNames, fileName)
	}
	sort.Strings(fileNames)
	return fileNames
}

func HasManifestFile(dir, manifestFileName string) bool {
	_, err := os.Stat(filepath.Join(dir, manifestFileName))
	return err == nil
}

var ScanIgnoreFiles = []string{
	".tmp",
	"tmp",
	".git",
	".angular",
	".cache",
	"cache",
	"vendor",
	"node_modules",
}

var IgnoreFiles = []string{
	".gitignore",
	".dockerignore",
}

func IsScanIgnored(name string) bool {
	for _, ignored := range ScanIgnoreFiles {
		if ignored == name {
			return true
		}
	}
	return false
}

func IsIgnoredFile(name string) bool {
	for _, ignored := range IgnoreFiles {
		if ignored == name {
			return true
		}
	}
	return false
}
