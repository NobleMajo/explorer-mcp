package parent

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
)

func listParentProjects(startDir string, maxDepth int) ([]string, error) {
	siblings := make([]string, 0)
	if maxDepth < 1 {
		return siblings, nil
	}

	cwd := filepath.Clean(startDir)
	seen := make(map[string]struct{})

	err := scanParentRepositories(cwd, maxDepth, func(absPath string) {
		rel, relErr := filepath.Rel(cwd, absPath)
		if relErr != nil {
			return
		}

		rel = filepath.ToSlash(rel)
		if rel == "." || !strings.HasPrefix(rel, "..") || !hasNamedPathSegment(rel) {
			return
		}
		if _, ok := seen[rel]; ok {
			return
		}
		seen[rel] = struct{}{}

		siblings = append(siblings, formatSiblingProject(absPath, rel))
	})
	if err != nil {
		return nil, err
	}

	return siblings, nil
}

func scanParentRepositories(startDir string, maxDepth int, callback func(path string)) error {
	current := filepath.Clean(startDir)
	previous := ""

	for p := 1; p <= maxDepth; p++ {
		parent := filepath.Dir(current)
		if parent == current {
			break
		}

		downDepth := maxDepth - p + 1
		if err := scanDownwards(parent, previous, downDepth, 1, callback); err != nil {
			return err
		}

		previous = current
		current = parent
	}

	return nil
}

func scanDownwards(currentDir, skipDir string, maxDepth, currentDepth int, callback func(path string)) error {
	if currentDepth > maxDepth {
		return nil
	}

	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return err
	}

	skipDir = filepath.Clean(skipDir)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if globals.IsScanIgnored(entry.Name()) {
			continue
		}

		path := filepath.Clean(filepath.Join(currentDir, entry.Name()))
		if path == skipDir {
			continue
		}

		callback(path)

		if err := scanDownwards(path, "", maxDepth, currentDepth+1, callback); err != nil {
			return err
		}
	}

	return nil
}

func hasNamedPathSegment(rel string) bool {
	for _, part := range strings.Split(rel, "/") {
		if part != ".." && part != "." && part != "" {
			return true
		}
	}
	return false
}
