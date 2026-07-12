package service

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func loadGoModManifest(root, manifestPath string) (ecosystemResult, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return ecosystemResult{}, err
	}

	deps := parseGoModRequire(string(data))
	return ecosystemResult{
		EcosystemName:        "go",
		ManifestFilePath:     "go.mod",
		ManifestFileExists:   true,
		IsParsed:             true,
		DependencyCount:      len(deps),
		Dependencies:         deps,
	}, nil
}

func loadPackageJsonManifest(root, manifestPath string) (ecosystemResult, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return ecosystemResult{}, err
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return ecosystemResult{}, err
	}

	groups := make([]dependencyGroup, 0, 2)
	if len(pkg.Dependencies) > 0 {
		groups = append(groups, dependencyGroup{
			GroupName:    "dependencies",
			PackageNames: sortedManifestKeys(pkg.Dependencies),
		})
	}
	if len(pkg.DevDependencies) > 0 {
		groups = append(groups, dependencyGroup{
			GroupName:    "devDependencies",
			PackageNames: sortedManifestKeys(pkg.DevDependencies),
		})
	}

	return ecosystemResult{
		EcosystemName:      "node",
		ManifestFilePath:   "package.json",
		ManifestFileExists: true,
		IsParsed:           true,
		DependencyGroups:   groups,
	}, nil
}

func loadRequirementsManifest(root, manifestPath string) (ecosystemResult, error) {
	file, err := os.Open(manifestPath)
	if err != nil {
		return ecosystemResult{}, err
	}
	defer file.Close()

	names := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		if line == "" {
			continue
		}
		name := line
		if before, _, ok := strings.Cut(line, "=="); ok {
			name = strings.TrimSpace(before)
		} else if before, _, ok := strings.Cut(line, ">="); ok {
			name = strings.TrimSpace(before)
		}
		names = append(names, name)
	}
	if err := scanner.Err(); err != nil {
		return ecosystemResult{}, err
	}

	sort.Strings(names)

	return ecosystemResult{
		EcosystemName:      "python",
		ManifestFilePath:   "requirements.txt",
		ManifestFileExists: true,
		IsParsed:           true,
		DependencyGroups: []dependencyGroup{{
			GroupName:    "requirements",
			PackageNames: names,
		}},
	}, nil
}

func loadCargoManifest(root, manifestPath string) (ecosystemResult, error) {
	return detectOnlyManifestResult(manifestPath, "full TOML parsing not in v1")
}

func loadPyprojectManifest(root, manifestPath string) (ecosystemResult, error) {
	return detectOnlyManifestResult(manifestPath, "full TOML parsing not in v1")
}

func detectOnlyManifestResult(manifestPath, reason string) (ecosystemResult, error) {
	return ecosystemResult{
		EcosystemName:      ecosystemNameForManifest(manifestPath),
		ManifestFilePath:   filepath.Base(manifestPath),
		ManifestFileExists: true,
		IsParsed:           false,
		ParseSkipReason:    reason,
	}, nil
}

func ecosystemNameForManifest(manifestPath string) string {
	switch filepath.Base(manifestPath) {
	case "go.mod":
		return "go"
	case "package.json":
		return "node"
	case "requirements.txt", "pyproject.toml":
		return "python"
	case "Cargo.toml":
		return "rust"
	default:
		return "unknown"
	}
}

func parseGoModRequire(content string) []goDependency {
	lines := strings.Split(content, "\n")
	inBlock := false
	deps := make([]goDependency, 0)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}

		if strings.HasPrefix(trimmed, "require (") {
			inBlock = true
			continue
		}
		if inBlock && trimmed == ")" {
			inBlock = false
			continue
		}

		if inBlock {
			if dep, ok := parseGoModRequireLine(trimmed); ok {
				deps = append(deps, dep)
			}
			continue
		}

		if strings.HasPrefix(trimmed, "require ") {
			if dep, ok := parseGoModRequireLine(strings.TrimPrefix(trimmed, "require ")); ok {
				deps = append(deps, dep)
			}
		}
	}

	sort.Slice(deps, func(i, j int) bool {
		return deps[i].PackageName < deps[j].PackageName
	})

	return deps
}

func parseGoModRequireLine(line string) (goDependency, bool) {
	isIndirect := strings.Contains(line, "// indirect")
	line = strings.Split(line, "//")[0]
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) < 2 {
		return goDependency{}, false
	}

	return goDependency{
		PackageName: fields[0],
		Version:     fields[1],
		IsIndirect:  isIndirect,
	}, true
}

func sortedManifestKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
