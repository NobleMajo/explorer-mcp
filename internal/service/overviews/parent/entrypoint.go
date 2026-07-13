package parent

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func ParentOverview(settings ScanSettings) func() resource.ExploreResource {
	return func() resource.ExploreResource {
		return func(projectRootPath string, verbose bool) (any, error) {
			return buildWorkspaceContext(projectRootPath, verbose, settings)
		}
	}
}
