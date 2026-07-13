package git

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func GitOverview() resource.OverviewResource {
	return func(verbose bool) (any, error) {
		return buildGitOverview(verbose)
	}
}
