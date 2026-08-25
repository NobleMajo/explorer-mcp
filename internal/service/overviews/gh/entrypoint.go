package gh

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func GhOverview() resource.ExploreResource {
	return func(projectRootPath string, verbose bool) (any, error) {
		return buildGhOverview(projectRootPath, verbose)
	}
}
