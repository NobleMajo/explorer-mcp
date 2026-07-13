package deps

import (
	"os"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
)

func buildDependencies(verbose bool) ([]string, error) {
	_ = verbose
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return globals.CollectManifestDependencies(root)
}
