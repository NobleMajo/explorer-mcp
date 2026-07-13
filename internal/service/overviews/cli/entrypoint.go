package cli

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func CLIOverview() resource.OverviewResource {
	return func(verbose bool) (any, error) {
		return buildCLIOverview(verbose)
	}
}
