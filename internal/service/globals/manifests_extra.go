package globals

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"sort"
	"strings"
)

var (
	trailingCommaPattern = regexp.MustCompile(`,(\s*[}\]])`)
	findPackagePattern   = regexp.MustCompile(`(?i)find_package\s*\(\s*([A-Za-z0-9_]+)`)
	fetchContentPattern  = regexp.MustCompile(`(?i)FetchContent_Declare\s*\(\s*([A-Za-z0-9_]+)`)
	gemPattern           = regexp.MustCompile(`^gem\s+['"]([^'"]+)['"]`)
)

func readManifestBytes(manifestPath string) ([]byte, bool, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return data, true, nil
}

func formatScopedDependency(name, version, scope string) string {
	if version == "" {
		if scope == "" {
			return name
		}
		return name + " " + scope
	}
	if scope == "" {
		return name + "@" + version
	}
	return name + "@" + version + " " + scope
}

func stripTrailingCommas(data []byte) []byte {
	return trailingCommaPattern.ReplaceAll(data, []byte("$1"))
}

func stripJSONComments(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if idx := strings.Index(line, "//"); idx >= 0 {
			before := strings.TrimSpace(line[:idx])
			if strings.Count(before, `"`)%2 == 0 {
				line = line[:idx]
			}
		}
		out = append(out, line)
	}
	return []byte(strings.Join(out, "\n"))
}

func parseLenientJSON(data []byte, v any) error {
	return json.Unmarshal(stripTrailingCommas(data), v)
}

func splitNameVersion(ref string) (string, string) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", ""
	}

	at := strings.LastIndex(ref, "@")
	if at <= 0 {
		return ref, ""
	}
	return ref[:at], ref[at+1:]
}

func parseDenoDependencyRef(name, ref, scope string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return formatScopedDependency(name, "", scope)
	}

	if strings.HasPrefix(ref, "npm:") {
		pkgName, version := splitNameVersion(strings.TrimPrefix(ref, "npm:"))
		return formatScopedDependency(pkgName, version, scope)
	}
	if strings.HasPrefix(ref, "jsr:") {
		pkgName, version := splitNameVersion(strings.TrimPrefix(ref, "jsr:"))
		return formatScopedDependency(pkgName, version, scope)
	}
	if strings.HasPrefix(ref, "https://") || strings.HasPrefix(ref, "http://") {
		return formatScopedDependency(name, ref, scope)
	}

	pkgName, version := splitNameVersion(ref)
	if pkgName == "" {
		pkgName = name
	}
	return formatScopedDependency(pkgName, version, scope)
}

func LoadBunLockManifest(root, manifestPath string) (string, []string, bool, error) {
	_ = root
	data, loaded, err := readManifestBytes(manifestPath)
	if err != nil || !loaded {
		return "", nil, loaded, err
	}

	var lock struct {
		Workspaces map[string]struct {
			Dependencies     map[string]string `json:"dependencies"`
			DevDependencies  map[string]string `json:"devDependencies"`
			PeerDependencies map[string]string `json:"peerDependencies"`
		} `json:"workspaces"`
		Packages map[string]json.RawMessage `json:"packages"`
	}
	if err := parseLenientJSON(data, &lock); err != nil {
		return "", nil, false, err
	}

	workspace := lock.Workspaces[""]
	if workspace.Dependencies == nil && workspace.DevDependencies == nil && workspace.PeerDependencies == nil {
		for _, candidate := range lock.Workspaces {
			workspace = candidate
			break
		}
	}

	entries := make([]string, 0)
	seen := make(map[string]struct{})

	add := func(name, version, scope string) {
		entry := formatScopedDependency(name, version, scope)
		if _, ok := seen[entry]; ok {
			return
		}
		seen[entry] = struct{}{}
		entries = append(entries, entry)
	}

	for name, version := range workspace.Dependencies {
		add(name, version, DepScopeDirect)
	}
	for name, version := range workspace.DevDependencies {
		add(name, version, DepScopeDev)
	}
	for name, version := range workspace.PeerDependencies {
		add(name, version, DepScopeDirect)
	}

	for pkgName, raw := range lock.Packages {
		var pkgEntry []json.RawMessage
		if err := json.Unmarshal(raw, &pkgEntry); err != nil || len(pkgEntry) == 0 {
			continue
		}
		var resolved string
		if err := json.Unmarshal(pkgEntry[0], &resolved); err != nil {
			continue
		}
		name, version := splitNameVersion(resolved)
		if name == "" {
			name = pkgName
		}
		if version == "" {
			continue
		}

		entry := formatScopedDependency(name, version, DepScopeIndirect)
		if _, ok := seen[entry]; ok {
			continue
		}
		directEntry := formatScopedDependency(name, version, DepScopeDirect)
		devEntry := formatScopedDependency(name, version, DepScopeDev)
		if _, ok := seen[directEntry]; ok {
			continue
		}
		if _, ok := seen[devEntry]; ok {
			continue
		}
		seen[entry] = struct{}{}
		entries = append(entries, entry)
	}

	sort.Strings(entries)
	return "@bun", entries, true, nil
}

func LoadCMakeManifest(root, manifestPath string) (string, []string, bool, error) {
	_ = root
	data, loaded, err := readManifestBytes(manifestPath)
	if err != nil || !loaded {
		return "", nil, loaded, err
	}

	seen := make(map[string]struct{})
	entries := make([]string, 0)
	content := string(data)

	for _, pattern := range []*regexp.Regexp{findPackagePattern, fetchContentPattern} {
		for _, match := range pattern.FindAllStringSubmatch(content, -1) {
			if len(match) < 2 {
				continue
			}
			name := match[1]
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			entries = append(entries, formatScopedDependency(name, "", DepScopeDirect))
		}
	}

	sort.Strings(entries)
	return "@cmake", entries, true, nil
}

func LoadDenoManifest(root, manifestPath string) (string, []string, bool, error) {
	_ = root
	data, loaded, err := readManifestBytes(manifestPath)
	if err != nil || !loaded {
		return "", nil, loaded, err
	}

	if strings.HasSuffix(strings.ToLower(manifestPath), ".jsonc") {
		data = stripJSONComments(data)
	}

	var manifest struct {
		Imports         map[string]string `json:"imports"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return "", nil, false, err
	}

	entries := make([]string, 0)
	seen := make(map[string]struct{})
	add := func(entry string) {
		if _, ok := seen[entry]; ok {
			return
		}
		seen[entry] = struct{}{}
		entries = append(entries, entry)
	}

	for name, ref := range manifest.Imports {
		add(parseDenoDependencyRef(name, ref, DepScopeDirect))
	}
	for name, ref := range manifest.Dependencies {
		add(parseDenoDependencyRef(name, ref, DepScopeDirect))
	}
	for name, ref := range manifest.DevDependencies {
		add(parseDenoDependencyRef(name, ref, DepScopeDev))
	}

	sort.Strings(entries)
	return "@deno", entries, true, nil
}

func LoadGoWorkManifest(root, manifestPath string) (string, []string, bool, error) {
	_ = root
	data, loaded, err := readManifestBytes(manifestPath)
	if err != nil || !loaded {
		return "", nil, loaded, err
	}

	entries := parseGoWorkUse(string(data))
	sort.Strings(entries)
	return "@go-workspace", entries, true, nil
}

func parseGoWorkUse(content string) []string {
	lines := strings.Split(content, "\n")
	inBlock := false
	entries := make([]string, 0)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}

		if strings.HasPrefix(trimmed, "use (") {
			inBlock = true
			continue
		}
		if inBlock && trimmed == ")" {
			inBlock = false
			continue
		}

		if inBlock {
			path := strings.TrimSpace(strings.Trim(trimmed, ","))
			if path != "" {
				entries = append(entries, formatScopedDependency(path, "", DepScopeDirect))
			}
			continue
		}

		if strings.HasPrefix(trimmed, "use ") {
			path := strings.TrimSpace(strings.TrimPrefix(trimmed, "use "))
			path = strings.Trim(path, ",")
			if path != "" {
				entries = append(entries, formatScopedDependency(path, "", DepScopeDirect))
			}
		}
	}

	return entries
}

func LoadComposerManifest(root, manifestPath string) (string, []string, bool, error) {
	_ = root
	data, loaded, err := readManifestBytes(manifestPath)
	if err != nil || !loaded {
		return "", nil, loaded, err
	}

	var manifest struct {
		Require    map[string]string `json:"require"`
		RequireDev map[string]string `json:"require-dev"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return "", nil, false, err
	}

	entries := make([]string, 0, len(manifest.Require)+len(manifest.RequireDev))
	for name, version := range manifest.Require {
		entries = append(entries, formatScopedDependency(name, version, DepScopeDirect))
	}
	for name, version := range manifest.RequireDev {
		entries = append(entries, formatScopedDependency(name, version, DepScopeDev))
	}

	sort.Strings(entries)
	return "@composer", entries, true, nil
}

func LoadGemfileManifest(root, manifestPath string) (string, []string, bool, error) {
	_ = root
	data, loaded, err := readManifestBytes(manifestPath)
	if err != nil || !loaded {
		return "", nil, loaded, err
	}

	entries := parseGemfile(string(data))
	sort.Strings(entries)
	return "@ruby", entries, true, nil
}

func parseGemfile(content string) []string {
	lines := strings.Split(content, "\n")
	scope := DepScopeDirect
	entries := make([]string, 0)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "group ") {
			if strings.Contains(lower, ":development") || strings.Contains(lower, ":test") || strings.Contains(lower, ":dev") {
				scope = DepScopeDev
			} else {
				scope = DepScopeDirect
			}
			continue
		}
		if trimmed == "end" {
			scope = DepScopeDirect
			continue
		}

		match := gemPattern.FindStringSubmatch(trimmed)
		if len(match) < 2 {
			continue
		}

		name := match[1]
		version := ""
		rest := strings.TrimSpace(strings.TrimPrefix(trimmed, match[0]))
		rest = strings.TrimPrefix(rest, ",")
		rest = strings.TrimSpace(rest)
		if strings.HasPrefix(rest, "'") || strings.HasPrefix(rest, `"`) {
			quote := rest[:1]
			rest = rest[1:]
			if end := strings.Index(rest, quote); end >= 0 {
				version = rest[:end]
			}
		}

		entries = append(entries, formatScopedDependency(name, version, scope))
	}

	return entries
}

func LoadPubspecManifest(root, manifestPath string) (string, []string, bool, error) {
	_ = root
	data, loaded, err := readManifestBytes(manifestPath)
	if err != nil || !loaded {
		return "", nil, loaded, err
	}

	entries := parsePubspecDependencies(string(data))
	sort.Strings(entries)
	return "@dart", entries, true, nil
}

func parsePubspecDependencies(content string) []string {
	lines := strings.Split(content, "\n")
	section := ""
	pendingName := ""
	entries := make([]string, 0)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		switch trimmed {
		case "dependencies:":
			section = DepScopeDirect
			pendingName = ""
			continue
		case "dev_dependencies:":
			section = DepScopeDev
			pendingName = ""
			continue
		}

		if section == "" {
			continue
		}

		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			section = ""
			pendingName = ""
			continue
		}

		if strings.HasSuffix(trimmed, ":") && !strings.Contains(strings.TrimSuffix(trimmed, ":"), " ") {
			pendingName = strings.TrimSuffix(trimmed, ":")
			continue
		}

		if pendingName != "" && strings.Contains(trimmed, ":") {
			parts := strings.SplitN(trimmed, ":", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			version := value
			if key == "sdk" {
				version = "sdk:" + value
			}
			entries = append(entries, formatScopedDependency(pendingName, version, section))
			pendingName = ""
			continue
		}

		pendingName = ""
		if !strings.Contains(trimmed, ":") {
			continue
		}

		parts := strings.SplitN(trimmed, ":", 2)
		name := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if name == "" || value == "" {
			continue
		}

		entries = append(entries, formatScopedDependency(name, value, section))
	}

	return entries
}
