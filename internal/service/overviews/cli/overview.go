package cli

import (
	"os/exec"
	"sort"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
)

type cliOverviewResponse struct {
	CommonCliToolsFound []string `json:"commonCliToolsFound,omitempty"`
}

func buildCLIOverview(verbose bool) (cliOverviewResponse, error) {
	_ = verbose

	found := make([]string, 0)
	for _, name := range globals.CommonCLIToolNames {
		if _, err := exec.LookPath(name); err == nil {
			found = append(found, name)
		}
	}

	sort.Strings(found)

	return cliOverviewResponse{
		CommonCliToolsFound: found,
	}, nil
}
