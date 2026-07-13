package structure

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func StructureOverview(repoScanDepth int) func() resource.OverviewResource {
	return func() resource.OverviewResource {
		return func(verbose bool) (any, error) {
			return buildRepoStructure(verbose, repoScanDepth)
		}
	}
}
