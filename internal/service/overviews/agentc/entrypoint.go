package agentc

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func AgentcOverview() resource.ExploreResource {
	return func(projectRootPath string, verbose bool) (any, error) {
		return buildAgentcOverview(projectRootPath, verbose)
	}
}
