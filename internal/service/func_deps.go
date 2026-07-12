package service

import (
	"os"
	"path/filepath"
	"sort"
)

type dependenciesResponse struct {
	responseMeta
	ProjectRootPath        string            `json:"projectRootPath"`
	DetectedEcosystemNames []string          `json:"detectedEcosystemNames"`
	EcosystemCount         int               `json:"ecosystemCount"`
	Ecosystems             []ecosystemResult `json:"ecosystems"`
}

type ecosystemResult struct {
	EcosystemName      string            `json:"ecosystemName"`
	ManifestFilePath   string            `json:"manifestFilePath"`
	ManifestFileExists bool              `json:"manifestFileExists"`
	IsParsed           bool              `json:"isParsed"`
	ParseSkipReason    string            `json:"parseSkipReason,omitempty"`
	DependencyCount    int               `json:"dependencyCount,omitempty"`
	Dependencies       []goDependency    `json:"dependencies,omitempty"`
	DependencyGroups   []dependencyGroup `json:"dependencyGroups,omitempty"`
}

type goDependency struct {
	PackageName string `json:"packageName"`
	Version     string `json:"version"`
	IsIndirect  bool   `json:"isIndirect"`
}

type dependencyGroup struct {
	GroupName    string   `json:"groupName"`
	PackageNames []string `json:"packageNames"`
}

func Dependencies() (string, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}

	ecosystems := make([]ecosystemResult, 0)
	names := make([]string, 0)

	fileNames := make([]string, 0, len(manifestLoaders))
	for fileName := range manifestLoaders {
		fileNames = append(fileNames, fileName)
	}
	sort.Strings(fileNames)

	for _, fileName := range fileNames {
		manifestPath := filepath.Join(root, fileName)
		if _, err := os.Stat(manifestPath); err != nil {
			continue
		}

		result, err := manifestLoaders[fileName](root, manifestPath)
		if err != nil {
			return "", err
		}
		ecosystems = append(ecosystems, result)
		if !containsEcosystemName(names, result.EcosystemName) {
			names = append(names, result.EcosystemName)
		}
	}

	sort.Strings(names)

	return marshalResponse(dependenciesResponse{
		responseMeta: responseMeta{
			ToolName:      "dependencies",
			SchemaVersion: schemaVersion,
		},
		ProjectRootPath:        root,
		DetectedEcosystemNames: names,
		EcosystemCount:         len(ecosystems),
		Ecosystems:             ecosystems,
	})
}

func containsEcosystemName(names []string, name string) bool {
	for _, existing := range names {
		if existing == name {
			return true
		}
	}
	return false
}
