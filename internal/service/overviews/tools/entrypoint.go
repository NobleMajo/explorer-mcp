package tools

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func ToolsOverview() resource.ExploreResource {
	return func(projectRootPath string, verbose bool) (any, error) {
		return buildProjectTools(projectRootPath, verbose)
	}
}
