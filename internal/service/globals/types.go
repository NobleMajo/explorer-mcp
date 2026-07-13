package globals

type EcosystemResult struct {
	EcosystemName      string            `json:"ecosystemName"`
	ManifestFilePath   string            `json:"manifestFilePath"`
	ManifestFileExists bool              `json:"manifestFileExists"`
	IsParsed           bool              `json:"isParsed"`
	ParseSkipReason    string            `json:"parseSkipReason,omitempty"`
	DependencyCount    int               `json:"dependencyCount,omitempty"`
	Dependencies       []GoDependency    `json:"dependencies,omitempty"`
	DependencyGroups   []DependencyGroup `json:"dependencyGroups,omitempty"`
}

type GoDependency struct {
	PackageName string `json:"packageName"`
	Version     string `json:"version"`
	IsIndirect  bool   `json:"isIndirect"`
}

type DependencyGroup struct {
	GroupName    string   `json:"groupName"`
	PackageNames []string `json:"packageNames"`
}
