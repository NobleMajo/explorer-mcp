package globals

import (
	"bufio"
	"encoding/json"
	"os"
	"sort"
	"strings"
)

func LoadGoModManifest(root, manifestPath string) ([]string, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	return parseGoModRequire(string(data)), nil
}

func LoadPackageJsonManifest(root, manifestPath string) ([]string, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	entries := make([]string, 0, len(pkg.Dependencies)+len(pkg.DevDependencies))
	for name, version := range pkg.Dependencies {
		entries = append(entries, formatNodeDependency(name, version, "production"))
	}
	for name, version := range pkg.DevDependencies {
		entries = append(entries, formatNodeDependency(name, version, "development"))
	}

	sort.Strings(entries)
	return entries, nil
}

func LoadRequirementsManifest(root, manifestPath string) ([]string, error) {
	file, err := os.Open(manifestPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	entries := make([]string, 0)
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
		entries = append(entries, formatPythonDependency(line))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	sort.Strings(entries)
	return entries, nil
}

func LoadCargoManifest(root, manifestPath string) ([]string, error) {
	_, err := os.Stat(manifestPath)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func LoadPyprojectManifest(root, manifestPath string) ([]string, error) {
	_, err := os.Stat(manifestPath)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func formatGoModuleDependency(name, version string, indirect bool) string {
	scope := "direct"
	if indirect {
		scope = "indirect"
	}
	return name + "@" + version + " " + scope
}

func formatNodeDependency(name, version, scope string) string {
	return name + "@" + version + " " + scope
}

func formatPythonDependency(line string) string {
	operators := []string{"==", ">=", "<=", "~=", "!=", ">", "<"}
	for _, op := range operators {
		if idx := strings.Index(line, op); idx > 0 {
			name := strings.TrimSpace(line[:idx])
			constraint := strings.TrimSpace(line[idx:])
			return name + "@" + constraint
		}
	}
	return strings.TrimSpace(line)
}

func parseGoModRequire(content string) []string {
	lines := strings.Split(content, "\n")
	inBlock := false
	entries := make([]string, 0)

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
			if entry, ok := parseGoModRequireLine(trimmed); ok {
				entries = append(entries, entry)
			}
			continue
		}

		if strings.HasPrefix(trimmed, "require ") {
			if entry, ok := parseGoModRequireLine(strings.TrimPrefix(trimmed, "require ")); ok {
				entries = append(entries, entry)
			}
		}
	}

	sort.Strings(entries)
	return entries
}

func parseGoModRequireLine(line string) (string, bool) {
	isIndirect := strings.Contains(line, "// indirect")
	line = strings.Split(line, "//")[0]
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) < 2 {
		return "", false
	}

	return formatGoModuleDependency(fields[0], fields[1], isIndirect), true
}

func sortedManifestKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
