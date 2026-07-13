package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/NobleMajo/explorer-mcp/internal/fsutil"
	"github.com/NobleMajo/explorer-mcp/internal/jsonresp"
)

type projectToolsResponse struct {
	jsonresp.Meta
	ProjectRootPath        string   `json:"projectRootPath"`
	HasMakefile            bool     `json:"hasMakefile"`
	MakefileTargetCount    int      `json:"makefileTargetCount"`
	MakefileTargetNames    []string `json:"makefileTargetNames"`
	HasPackageJson         bool     `json:"hasPackageJson"`
	PackageJsonScriptCount int      `json:"packageJsonScriptCount"`
	PackageJsonScriptNames []string `json:"packageJsonScriptNames"`
	ShellScriptCount       int      `json:"shellScriptCount"`
	ShellScriptPaths       []string `json:"shellScriptPaths"`
}

func buildProjectTools(verbose bool) (projectToolsResponse, error) {
	_ = verbose
	root, err := os.Getwd()
	if err != nil {
		return projectToolsResponse{}, err
	}

	resp := projectToolsResponse{
		Meta: jsonresp.Meta{
			ToolName:      "project_tools",
			SchemaVersion: jsonresp.SchemaVersion,
		},
		ProjectRootPath:        root,
		MakefileTargetNames:    []string{},
		PackageJsonScriptNames: []string{},
		ShellScriptPaths:       []string{},
	}

	makefilePath := filepath.Join(root, "Makefile")
	if fsutil.FileExists(makefilePath) {
		resp.HasMakefile = true
		data, err := os.ReadFile(makefilePath)
		if err != nil {
			return projectToolsResponse{}, err
		}
		resp.MakefileTargetNames = parseMakefileTargetNames(string(data))
		resp.MakefileTargetCount = len(resp.MakefileTargetNames)
	}

	packageJSONPath := filepath.Join(root, "package.json")
	if fsutil.FileExists(packageJSONPath) {
		resp.HasPackageJson = true
		names, err := parsePackageJsonScriptNames(packageJSONPath)
		if err != nil {
			return projectToolsResponse{}, err
		}
		resp.PackageJsonScriptNames = names
		resp.PackageJsonScriptCount = len(names)
	}

	shellScripts, err := listRootShellScripts(root)
	if err != nil {
		return projectToolsResponse{}, err
	}
	resp.ShellScriptPaths = shellScripts
	resp.ShellScriptCount = len(shellScripts)

	return resp, nil
}

func parseMakefileTargetNames(content string) []string {
	targets := make([]string, 0)
	seen := make(map[string]bool)

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "\t") {
			continue
		}
		if strings.HasPrefix(line, ".") {
			continue
		}

		before, _, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		for _, part := range strings.Fields(before) {
			if part == "" || seen[part] {
				continue
			}
			seen[part] = true
			targets = append(targets, part)
		}
	}

	sort.Strings(targets)
	return targets
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
