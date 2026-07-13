package deps

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func DepsOverview(settings Settings) func() resource.ExploreResource {
	return func() resource.ExploreResource {
		return func(projectRootPath string, verbose bool) (any, error) {
			return buildDependencies(projectRootPath, settings, verbose)
		}
	}
}
