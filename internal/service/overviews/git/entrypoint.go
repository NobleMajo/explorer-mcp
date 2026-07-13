package git

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func GitOverview(recentCommitCount int) func() resource.OverviewResource {
	return func() resource.OverviewResource {
		return func(verbose bool) (any, error) {
			return buildGitOverview(verbose, recentCommitCount)
		}
	}
}
