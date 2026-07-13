package globals

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

type ManifestLoader func(root, manifestPath string) (EcosystemResult, error)

var ManifestLoaders = map[string]ManifestLoader{
	"go.mod":           LoadGoModManifest,
	"package.json":     LoadPackageJsonManifest,
	"requirements.txt": LoadRequirementsManifest,
	"Cargo.toml":       LoadCargoManifest,
	"pyproject.toml":   LoadPyprojectManifest,
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

const StructureScanMaxDepth = 3

var FollowGitIgnore bool = true
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
