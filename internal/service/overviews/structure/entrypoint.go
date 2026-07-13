package structure

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func StructureOverview() resource.OverviewResource {
	return func(verbose bool) (any, error) {
		return buildRepoStructure(verbose)
	}
}
