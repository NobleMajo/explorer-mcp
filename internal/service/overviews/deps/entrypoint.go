package deps

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func DepsOverview() resource.OverviewResource {
	return func(verbose bool) (any, error) {
		return buildDependencies(verbose)
	}
}
