package structure

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
)

type repoStructureResponse struct {
	RepoScanDepthLimit *int     `json:"repoScanDepthLimit,omitempty"`
	EntryCount         *int     `json:"entryCount,omitempty"`
	Entries            []string `json:"entries,omitempty"`
}

func buildRepoStructure(verbose bool, repoScanDepth int) (repoStructureResponse, error) {
	_ = verbose
	if repoScanDepth < 1 {
		zero := 0
		return repoStructureResponse{RepoScanDepthLimit: &zero}, nil
	}

	root, err := os.Getwd()
	if err != nil {
		return repoStructureResponse{}, err
	}

	entries := make([]string, 0)
	if err := appendStructureEntries(root, root, 0, repoScanDepth, &entries); err != nil {
		return repoStructureResponse{}, err
	}

	count := len(entries)
	resp := repoStructureResponse{
		RepoScanDepthLimit: &repoScanDepth,
		EntryCount:         &count,
	}
	if len(entries) > 0 {
		resp.Entries = entries
	}

	return resp, nil
}

func appendStructureEntries(root, dir string, depth, maxDepth int, entries *[]string) error {
	if depth >= maxDepth {
		return nil
	}

	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	sort.Slice(dirEntries, func(i, j int) bool {
		return dirEntries[i].Name() < dirEntries[j].Name()
	})

	for _, entry := range dirEntries {
		if globals.IsScanIgnored(entry.Name()) {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())
		relPath, err := filepath.Rel(root, fullPath)
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if depth+1 >= maxDepth {
				hasMore, err := hasVisibleDescendants(fullPath)
				if err != nil {
					return err
				}
				if hasMore {
					*entries = append(*entries, filepath.ToSlash(relPath)+"/**")
				}
				continue
			}
			if err := appendStructureEntries(root, fullPath, depth+1, maxDepth, entries); err != nil {
				return err
			}
			continue
		}

		if globals.IsIgnoredFile(entry.Name()) {
			continue
		}

		*entries = append(*entries, filepath.ToSlash(relPath))
	}

	return nil
}

func hasVisibleDescendants(dir string) (bool, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	for _, entry := range dirEntries {
		if globals.IsScanIgnored(entry.Name()) {
			continue
		}

		if !entry.IsDir() {
			if globals.IsIgnoredFile(entry.Name()) {
				continue
			}
			return true, nil
		}

		childPath := filepath.Join(dir, entry.Name())
		hasMore, err := hasVisibleDescendants(childPath)
		if err != nil {
			return false, err
		}
		if hasMore {
			return true, nil
		}
	}

	return false, nil
}
