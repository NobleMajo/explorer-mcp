package opencode

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func OpencodeOverview() resource.OverviewResource {
	return func(projectRootPath string, verbose bool) (any, error) {
		return buildOpencodeOverview(projectRootPath, verbose)
	}
}
