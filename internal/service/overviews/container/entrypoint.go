package container

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func ContainerOverview() resource.OverviewResource {
	return func(projectRootPath string, verbose bool) (any, error) {
		return buildContainerOverview(projectRootPath, verbose)
	}
}
