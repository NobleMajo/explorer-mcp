package deps

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/NobleMajo/explorer-mcp/internal/jsonresp"
	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
)

type dependenciesResponse struct {
	jsonresp.Meta
	ProjectRootPath        string                    `json:"projectRootPath"`
	DetectedEcosystemNames []string                  `json:"detectedEcosystemNames"`
	EcosystemCount         int                       `json:"ecosystemCount"`
	Ecosystems             []globals.EcosystemResult `json:"ecosystems"`
}

func buildDependencies(verbose bool) (dependenciesResponse, error) {
	_ = verbose
	root, err := os.Getwd()
	if err != nil {
		return dependenciesResponse{}, err
	}

	ecosystems := make([]globals.EcosystemResult, 0)
	names := make([]string, 0)

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

		result, err := globals.ManifestLoaders[fileName](root, manifestPath)
		if err != nil {
			return dependenciesResponse{}, err
		}
		ecosystems = append(ecosystems, result)
		if !containsEcosystemName(names, result.EcosystemName) {
			names = append(names, result.EcosystemName)
		}
	}

	sort.Strings(names)

	return dependenciesResponse{
		Meta: jsonresp.Meta{
			ToolName:      "dependencies",
			SchemaVersion: jsonresp.SchemaVersion,
		},
		ProjectRootPath:        root,
		DetectedEcosystemNames: names,
		EcosystemCount:         len(ecosystems),
		Ecosystems:             ecosystems,
	}, nil
}

func containsEcosystemName(names []string, name string) bool {
	for _, existing := range names {
		if existing == name {
			return true
		}
	}
	return false
}
