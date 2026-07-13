package parent

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func ParentOverview(settings ScanSettings) func() resource.OverviewResource {
	return func() resource.OverviewResource {
		return func(verbose bool) (any, error) {
			return buildWorkspaceContext(verbose, settings)
		}
	}
}
