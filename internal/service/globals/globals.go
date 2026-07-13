package globals

import (
	"path/filepath"
	"slices"
	"sort"
	"strings"
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

type ManifestLoader func(root, manifestPath string) (string, []string, bool, error)

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

func CollectManifestDependencies(root string) ([]string, error) {
	dependencies := make([]string, 0)
	for _, fileName := range ManifestLoaderFileNames() {
		manifestPath := filepath.Join(root, fileName)
		_, entries, loaded, err := ManifestLoaders[fileName](root, manifestPath)
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

type ProjectIdentifier func(path string, subfiles []string, subdirs []string) ([]string, error)

var ProjectIdentifiers = []ProjectIdentifier{
	identifyMakefileProject,
	identifyTSConfigProject,
	identifyAngularProject,
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

func identifyMakefileProject(path string, subfiles []string, subdirs []string) ([]string, error) {
	_ = path
	_ = subdirs
	if containsFileName(subfiles, "Makefile") {
		return []string{"@makefile"}, nil
	}
	return nil, nil
}

func identifyTSConfigProject(path string, subfiles []string, subdirs []string) ([]string, error) {
	_ = path
	_ = subdirs
	for _, name := range subfiles {
		if isTSConfigFileName(name) {
			return []string{"@tsconfig"}, nil
		}
	}
	return nil, nil
}

func identifyAngularProject(path string, subfiles []string, subdirs []string) ([]string, error) {
	_ = path
	_ = subdirs
	if containsFileName(subfiles, "angular.json") {
		return []string{"@angular"}, nil
	}
	return nil, nil
}

func containsFileName(subfiles []string, name string) bool {
	for _, fileName := range subfiles {
		if fileName == name {
			return true
		}
	}
	return false
}

func isTSConfigFileName(name string) bool {
	if name == "tsconfig.json" {
		return true
	}
	if strings.HasSuffix(name, ".tsconfig.json") {
		return true
	}
	return strings.HasPrefix(name, "tsconfig.") && strings.HasSuffix(name, ".json")
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
