package cli

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func CLIOverview() resource.ExploreResource {
	return func(projectRootPath string, verbose bool) (any, error) {
		_ = projectRootPath
		return buildCLIOverview(verbose)
	}
}
