package parent

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func ParentOverview(parentScanDepth int) func() resource.OverviewResource {
	return func() resource.OverviewResource {
		return func(verbose bool) (any, error) {
			return buildWorkspaceContext(verbose, parentScanDepth)
		}
	}
}
