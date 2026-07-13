package container

import "github.com/NobleMajo/explorer-mcp/internal/service/resource"

func ContainerOverview() resource.OverviewResource {
	return func(verbose bool) (any, error) {
		return buildContainerOverview(verbose)
	}
}
