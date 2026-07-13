package deps

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
)

func buildDependencies(verbose bool) ([]string, error) {
	_ = verbose
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	dependencies := make([]string, 0)

	fileNames := make([]string, 0, len(globals.ManifestLoaders))
	for fileName := range globals.ManifestLoaders {
		fileNames = append(fileNames, fileName)
	}
	sort.Strings(fileNames)

	for _, fileName := range fileNames {
		manifestPath := filepath.Join(root, fileName)
		if _, err := os.Stat(manifestPath); err != nil {
			continue
		}

		entries, err := globals.ManifestLoaders[fileName](root, manifestPath)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, entries...)
	}

	sort.Strings(dependencies)
	return dependencies, nil
}
