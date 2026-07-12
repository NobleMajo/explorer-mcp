package service

type manifestLoader func(root, manifestPath string) (ecosystemResult, error)

var manifestLoaders = map[string]manifestLoader{
	"go.mod":           loadGoModManifest,
	"package.json":     loadPackageJsonManifest,
	"requirements.txt": loadRequirementsManifest,
	"Cargo.toml":       loadCargoManifest,
	"pyproject.toml":   loadPyprojectManifest,
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

const structureScanMaxDepth = 3

var IgnoreFiles = []string{
	".gitignore",
	".dockerignore",
}

func isScanIgnored(name string) bool {
	for _, ignored := range ScanIgnoreFiles {
		if ignored == name {
			return true
		}
	}
	return false
}

func isIgnoredFile(name string) bool {
	for _, ignored := range IgnoreFiles {
		if ignored == name {
			return true
		}
	}
	return false
}
