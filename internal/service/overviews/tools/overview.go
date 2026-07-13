package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/NobleMajo/explorer-mcp/internal/fsutil"
)

var makefileTargetNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`)

const (
	scriptsKeyMakeTargets      = "make-targets"
	scriptsKeyPackageJSON      = "package.json-scripts"
	scriptsKeyBashScripts      = "bash-scripts"
)

type projectToolsResponse struct {
	ToolsFound   []string            `json:"toolsFound,omitempty"`
	ScriptsFound map[string][]string `json:"scriptsFound,omitempty"`
}

func buildProjectTools(projectRootPath string, verbose bool) (projectToolsResponse, error) {
	_ = verbose
	root := projectRootPath

	toolsFound := make([]string, 0, 3)
	scriptsFound := make(map[string][]string)

	makefilePath := filepath.Join(root, "Makefile")
	if fsutil.FileExists(makefilePath) {
		toolsFound = append(toolsFound, "Makefile")
		data, err := os.ReadFile(makefilePath)
		if err != nil {
			return projectToolsResponse{}, err
		}
		if names := parseMakefileTargetNames(string(data)); len(names) > 0 {
			scriptsFound[scriptsKeyMakeTargets] = names
		}
	}

	packageJSONPath := filepath.Join(root, "package.json")
	if fsutil.FileExists(packageJSONPath) {
		toolsFound = append(toolsFound, "package.json")
		names, err := parsePackageJsonScriptNames(packageJSONPath)
		if err != nil {
			return projectToolsResponse{}, err
		}
		if len(names) > 0 {
			scriptsFound[scriptsKeyPackageJSON] = names
		}
	}

	shellScripts, err := listRootShellScripts(root)
	if err != nil {
		return projectToolsResponse{}, err
	}
	if len(shellScripts) > 0 {
		toolsFound = append(toolsFound, "*.sh")
		scriptsFound[scriptsKeyBashScripts] = shellScripts
	}

	sort.Strings(toolsFound)

	return projectToolsResponse{
		ToolsFound:   toolsFound,
		ScriptsFound: scriptsFound,
	}, nil
}

func parseMakefileTargetNames(content string) []string {
	targets := make([]string, 0)
	seen := make(map[string]bool)

	addTarget := func(name string) {
		if name == "" || seen[name] || !isValidMakefileTargetName(name) {
			return
		}
		seen[name] = true
		targets = append(targets, name)
	}

	for _, line := range strings.Split(content, "\n") {
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "\t") {
			continue
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.HasPrefix(trimmed, ".") {
			if phony, ok := strings.CutPrefix(trimmed, ".PHONY:"); ok {
				for _, part := range strings.Fields(phony) {
					addTarget(part)
				}
			}
			continue
		}

		if isMakefileDirective(trimmed) || isMakefileAssignment(trimmed) {
			continue
		}

		if strings.HasPrefix(trimmed, "@") || strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "+") {
			continue
		}

		before, _, ok := strings.Cut(trimmed, ":")
		if !ok {
			continue
		}

		for _, part := range strings.Fields(strings.TrimSpace(before)) {
			addTarget(part)
		}
	}

	sort.Strings(targets)
	return targets
}

func isMakefileAssignment(line string) bool {
	if strings.Contains(line, ":=") || strings.Contains(line, "?=") || strings.Contains(line, "+=") {
		return true
	}

	eq := strings.Index(line, "=")
	colon := strings.Index(line, ":")
	return eq >= 0 && (colon < 0 || eq < colon)
}

func isMakefileDirective(line string) bool {
	directive, _, _ := strings.Cut(line, " ")
	switch directive {
	case "ifdef", "ifndef", "endif", "else", "export", "include", "-include", "override", "define", "endef", "vpath", "unexport":
		return true
	default:
		return strings.HasPrefix(line, "$(shell")
	}
}

func isValidMakefileTargetName(name string) bool {
	return makefileTargetNamePattern.MatchString(name)
}

func parsePackageJsonScriptNames(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pkg struct {
		Scripts map[string]json.RawMessage `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	names := make([]string, 0, len(pkg.Scripts))
	for name := range pkg.Scripts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

func listRootShellScripts(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".sh") {
			paths = append(paths, name)
		}
	}

	sort.Strings(paths)
	return paths, nil
}
