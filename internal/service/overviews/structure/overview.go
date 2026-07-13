package structure

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
)

type repoStructureResponse struct {
	RepoScanPerformed bool     `json:"repoScanPerformed"`
	EntryCount        *int     `json:"entryCount,omitempty"`
	Entries           []string `json:"entries,omitempty"`
}

func buildRepoStructure(verbose bool, repoScanDepth int) (repoStructureResponse, error) {
	_ = verbose
	resp := repoStructureResponse{
		RepoScanPerformed: repoScanDepth > 0,
	}
	if repoScanDepth < 1 {
		return resp, nil
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
	resp.EntryCount = &count
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
