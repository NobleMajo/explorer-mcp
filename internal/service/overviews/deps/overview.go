package deps

import "github.com/NobleMajo/explorer-mcp/internal/service/globals"

func buildDependencies(projectRootPath string, verbose bool) ([]string, error) {
	_ = verbose
	return globals.CollectManifestDependencies(projectRootPath)
}
