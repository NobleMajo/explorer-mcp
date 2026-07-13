package globals

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"sort"
	"strings"
)

func LoadGoModManifest(root, manifestPath string) (string, []string, bool, error) {
	return loadGoModManifest(root, manifestPath, DefaultManifestDepsSettings())
}

func loadGoModManifest(root, manifestPath string, settings ManifestDepsSettings) (string, []string, bool, error) {
	_ = root
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil, false, nil
		}
		return "", nil, false, err
	}

	return "@go", parseGoModDependencies(string(data), settings), true, nil
}

func LoadPackageJsonManifest(root, manifestPath string) (string, []string, bool, error) {
	_ = root
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil, false, nil
		}
		return "", nil, false, err
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", nil, false, err
	}

	entries := make([]string, 0, len(pkg.Dependencies)+len(pkg.DevDependencies))
	for name, version := range pkg.Dependencies {
		entries = append(entries, formatScopedDependency(name, version, DepScopeDirect))
	}
	for name, version := range pkg.DevDependencies {
		entries = append(entries, formatScopedDependency(name, version, DepScopeDev))
	}

	sort.Strings(entries)
	return "@npm", entries, true, nil
}

func LoadRequirementsManifest(root, manifestPath string) (string, []string, bool, error) {
	_ = root
	file, err := os.Open(manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil, false, nil
		}
		return "", nil, false, err
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
		return "", nil, false, err
	}

	sort.Strings(entries)
	return "@pip", entries, true, nil
}

func LoadCargoManifest(root, manifestPath string) (string, []string, bool, error) {
	_ = root
	if _, err := os.Stat(manifestPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil, false, nil
		}
		return "", nil, false, err
	}
	return "@cargo", nil, true, nil
}

func LoadPyprojectManifest(root, manifestPath string) (string, []string, bool, error) {
	_ = root
	if _, err := os.Stat(manifestPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil, false, nil
		}
		return "", nil, false, err
	}
	return "@python", nil, true, nil
}

func formatGoModuleDependency(name, version string, indirect bool) string {
	scope := DepScopeDirect
	if indirect {
		scope = DepScopeIndirect
	}
	return formatScopedDependency(name, version, scope)
}

func formatPythonDependency(line string) string {
	operators := []string{"==", ">=", "<=", "~=", "!=", ">", "<"}
	for _, op := range operators {
		if idx := strings.Index(line, op); idx > 0 {
			name := strings.TrimSpace(line[:idx])
			constraint := strings.TrimSpace(line[idx:])
			return formatScopedDependency(name, constraint, DepScopeDirect)
		}
	}
	return formatScopedDependency(strings.TrimSpace(line), "", DepScopeDirect)
}

func parseGoModDependencies(content string, settings ManifestDepsSettings) []string {
	entries := parseGoModRequire(content)
	if !settings.ShowGoToolDeps {
		return entries
	}

	requireVersions := parseGoModRequireVersions(content)
	toolEntries := parseGoModTool(content, requireVersions)
	if len(toolEntries) == 0 {
		return entries
	}

	seen := make(map[string]struct{}, len(entries)+len(toolEntries))
	merged := make([]string, 0, len(entries)+len(toolEntries))
	for _, entry := range entries {
		if _, ok := seen[entry]; ok {
			continue
		}
		seen[entry] = struct{}{}
		merged = append(merged, entry)
	}
	for _, entry := range toolEntries {
		if _, ok := seen[entry]; ok {
			continue
		}
		seen[entry] = struct{}{}
		merged = append(merged, entry)
	}

	sort.Strings(merged)
	return merged
}

func parseGoModRequireVersions(content string) map[string]string {
	lines := strings.Split(content, "\n")
	inBlock := false
	versions := make(map[string]string)

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
			if mod, version, ok := parseGoModRequireFields(trimmed); ok {
				versions[mod] = version
			}
			continue
		}

		if strings.HasPrefix(trimmed, "require ") {
			if mod, version, ok := parseGoModRequireFields(strings.TrimPrefix(trimmed, "require ")); ok {
				versions[mod] = version
			}
		}
	}

	return versions
}

func parseGoModTool(content string, requireVersions map[string]string) []string {
	lines := strings.Split(content, "\n")
	inBlock := false
	entries := make([]string, 0)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}

		if strings.HasPrefix(trimmed, "tool (") {
			inBlock = true
			continue
		}
		if inBlock && trimmed == ")" {
			inBlock = false
			continue
		}

		if inBlock {
			if entry, ok := parseGoModToolLine(trimmed, requireVersions); ok {
				entries = append(entries, entry)
			}
			continue
		}

		if strings.HasPrefix(trimmed, "tool ") {
			if entry, ok := parseGoModToolLine(strings.TrimPrefix(trimmed, "tool "), requireVersions); ok {
				entries = append(entries, entry)
			}
		}
	}

	sort.Strings(entries)
	return entries
}

func parseGoModToolLine(line string, requireVersions map[string]string) (string, bool) {
	line = strings.Split(line, "//")[0]
	toolPath := strings.TrimSpace(strings.Trim(line, ","))
	if toolPath == "" {
		return "", false
	}

	version := lookupGoModToolVersion(toolPath, requireVersions)
	return formatScopedDependency(toolPath, version, DepScopeTool), true
}

func lookupGoModToolVersion(toolPath string, requireVersions map[string]string) string {
	bestModule := ""
	for modulePath, version := range requireVersions {
		if toolPath == modulePath || strings.HasPrefix(toolPath, modulePath+"/") {
			if len(modulePath) > len(bestModule) {
				bestModule = modulePath
				_ = version
			}
		}
	}
	if bestModule == "" {
		return ""
	}
	return requireVersions[bestModule]
}

func parseGoModRequireFields(line string) (string, string, bool) {
	line = strings.Split(line, "//")[0]
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) < 2 {
		return "", "", false
	}
	return fields[0], fields[1], true
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
