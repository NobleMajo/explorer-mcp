package tools

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func ToolsOverview() resource.OverviewResource {
	return func(verbose bool) (any, error) {
		return buildProjectTools(verbose)
	}
}
