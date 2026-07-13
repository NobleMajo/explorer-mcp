package structure

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func StructureOverview(settings ScanSettings) func() resource.ExploreResource {
	return func() resource.ExploreResource {
		return func(projectRootPath string, verbose bool) (any, error) {
			return buildRepoStructure(projectRootPath, verbose, settings)
		}
	}
}
