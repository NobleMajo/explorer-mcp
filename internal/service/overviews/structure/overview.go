package structure

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
)

type ScanSettings struct {
	Depth       int
	OutDirs     bool
	DepsDirs    bool
}

type repoStructureResponse struct {
	ProjectScanDepthLimit *int     `json:"projectScanDepthLimit,omitempty"`
	EntryCount            *int     `json:"entryCount,omitempty"`
	Entries               []string `json:"entries,omitempty"`
}

func buildRepoStructure(projectRootPath string, verbose bool, settings ScanSettings) (repoStructureResponse, error) {
	_ = verbose
	if settings.Depth < 1 {
		zero := 0
		return repoStructureResponse{ProjectScanDepthLimit: &zero}, nil
	}

	entries := make([]string, 0)
	if err := appendStructureEntries(projectRootPath, projectRootPath, 0, settings, &entries); err != nil {
		return repoStructureResponse{}, err
	}

	count := len(entries)
	resp := repoStructureResponse{
		ProjectScanDepthLimit: &settings.Depth,
		EntryCount:            &count,
	}
	if len(entries) > 0 {
		resp.Entries = entries
	}

	return resp, nil
}

func appendStructureEntries(root, dir string, depth int, settings ScanSettings, entries *[]string) error {
	if depth >= settings.Depth {
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
		fullPath := filepath.Join(dir, entry.Name())
		relPath, err := filepath.Rel(root, fullPath)
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if isOutputDir(entry.Name()) && !settings.OutDirs {
				*entries = append(*entries, filepath.ToSlash(relPath)+"/**")
				continue
			}
			if isDepsDir(entry.Name()) && !settings.DepsDirs {
				*entries = append(*entries, filepath.ToSlash(relPath)+"/**")
				continue
			}
		}

		if globals.IsScanIgnored(entry.Name()) {
			if !(entry.IsDir() && settings.DepsDirs && isDepsDir(entry.Name())) {
				continue
			}
		}
		if entry.IsDir() && isDotDir(entry.Name()) {
			continue
		}

		if entry.IsDir() {
			if depth+1 >= settings.Depth {
				hasMore, err := hasVisibleDescendants(fullPath, settings)
				if err != nil {
					return err
				}
				if hasMore {
					*entries = append(*entries, filepath.ToSlash(relPath)+"/**")
				}
				continue
			}
			if err := appendStructureEntries(root, fullPath, depth+1, settings, entries); err != nil {
				return err
			}
			continue
		}

		*entries = append(*entries, filepath.ToSlash(relPath))
	}

	return nil
}

func hasVisibleDescendants(dir string, settings ScanSettings) (bool, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	for _, entry := range dirEntries {
		if entry.IsDir() {
			if isOutputDir(entry.Name()) && !settings.OutDirs {
				return true, nil
			}
			if isDepsDir(entry.Name()) && !settings.DepsDirs {
				return true, nil
			}
		}

		if globals.IsScanIgnored(entry.Name()) {
			if !(entry.IsDir() && settings.DepsDirs && isDepsDir(entry.Name())) {
				continue
			}
		}
		if entry.IsDir() && isDotDir(entry.Name()) {
			continue
		}

		if !entry.IsDir() {
			return true, nil
		}

		childPath := filepath.Join(dir, entry.Name())
		hasMore, err := hasVisibleDescendants(childPath, settings)
		if err != nil {
			return false, err
		}
		if hasMore {
			return true, nil
		}
	}

	return false, nil
}

func isDotDir(name string) bool {
	return strings.HasPrefix(name, ".")
}

func isOutputDir(name string) bool {
	switch name {
	case "dist", "out", "output":
		return true
	default:
		return false
	}
}

func isDepsDir(name string) bool {
	switch name {
	case "node_modules", "vendor":
		return true
	default:
		return false
	}
}
