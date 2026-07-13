package parent

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func ParentOverview() resource.OverviewResource {
	return func(verbose bool) (any, error) {
		return buildWorkspaceContext(verbose)
	}
}
