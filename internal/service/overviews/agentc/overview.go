package agentc

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/NobleMajo/explorer-mcp/internal/fsutil"
)

// Known root-level instruction files checked by existence only (no content reads).
var rootInstructionFiles = []string{
	"AGENTS.md",
	"AGENT.md",
	"CLAUDE.md",
	".cursorrules",
	"CONTRIBUTING.md",
}

func buildAgentcOverview(projectRootPath string, verbose bool) (any, error) {
	_ = verbose

	found := make([]string, 0)
	seen := make(map[string]struct{})

	add := func(rel string) {
		rel = filepath.ToSlash(rel)
		if rel == "" {
			return
		}
		if _, ok := seen[rel]; ok {
			return
		}
		seen[rel] = struct{}{}
		found = append(found, rel)
	}

	for _, name := range rootInstructionFiles {
		if fsutil.FileExists(filepath.Join(projectRootPath, name)) {
			add(name)
		}
	}

	rulesDir := filepath.Join(projectRootPath, ".cursor", "rules")
	if fsutil.DirExists(rulesDir) {
		entries, err := os.ReadDir(rulesDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				name := entry.Name()
				// Skip OS/editor noise such as .DS_Store; keep only visible flat files.
				if strings.HasPrefix(name, ".") {
					continue
				}
				add(filepath.Join(".cursor", "rules", name))
			}
		}
		// Unreadable rules dir is treated as unavailable for that subtree.
	}

	docsDir := filepath.Join(projectRootPath, "docs")
	if fsutil.DirExists(docsDir) {
		entries, err := os.ReadDir(docsDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				name := entry.Name()
				if strings.HasPrefix(name, ".") {
					continue
				}
				lower := strings.ToLower(name)
				if strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".mdx") {
					add(filepath.Join("docs", name))
				}
			}
		}
		// Unreadable docs dir is treated as unavailable for that subtree.
	}

	if len(found) == 0 {
		return nil, nil
	}

	sort.Strings(found)
	return found, nil
}
