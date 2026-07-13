package structure

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/NobleMajo/explorer-mcp/internal/gitignore"
	"github.com/NobleMajo/explorer-mcp/internal/jsonresp"
	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
)

type repoStructureResponse struct {
	jsonresp.Meta
	RootPath    string           `json:"rootPath"`
	EntryCount  int              `json:"entryCount"`
	Entries     []structureEntry `json:"entries"`
}

type structureEntry struct {
	RelativePath string `json:"relativePath"`
	EntryName    string `json:"entryName"`
	IsDirectory  bool   `json:"isDirectory"`
	Depth        int    `json:"depth"`
}

type scanState struct {
	matchers []gitignore.DirMatcher
}

func newScanState() *scanState {
	return &scanState{
		matchers: make([]gitignore.DirMatcher, 0),
	}
}

func (s *scanState) loadMatcherForDir(dir, root string) error {
	if !globals.FollowGitIgnore {
		return nil
	}

	matcher, err := gitignore.LoadDirMatcher(dir, root)
	if err != nil {
		return err
	}
	if matcher != nil {
		s.matchers = append(s.matchers, *matcher)
	}
	return nil
}

func buildRepoStructure(verbose bool) (repoStructureResponse, error) {
	_ = verbose
	root, err := os.Getwd()
	if err != nil {
		return repoStructureResponse{}, err
	}

	entries := make([]structureEntry, 0)
	state := newScanState()
	if err := appendStructureEntries(root, root, 0, &entries, state); err != nil {
		return repoStructureResponse{}, err
	}

	return repoStructureResponse{
		Meta: jsonresp.Meta{
			ToolName:      "repo_structure",
			SchemaVersion: jsonresp.SchemaVersion,
		},
		RootPath:   root,
		EntryCount: len(entries),
		Entries:    entries,
	}, nil
}

func appendStructureEntries(root, dir string, depth int, entries *[]structureEntry, state *scanState) error {
	if depth >= globals.StructureScanMaxDepth {
		return nil
	}

	if err := state.loadMatcherForDir(dir, root); err != nil {
		return err
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
		if globals.FollowGitIgnore && gitignore.ShouldIgnore(state.matchers, relPath, entry.IsDir()) {
			continue
		}

		if entry.IsDir() {
			if err := appendStructureEntries(root, fullPath, depth+1, entries, state); err != nil {
				return err
			}
			continue
		}

		if globals.IsIgnoredFile(entry.Name()) {
			continue
		}

		*entries = append(*entries, structureEntry{
			RelativePath: filepath.ToSlash(relPath),
			EntryName:    entry.Name(),
			IsDirectory:  false,
			Depth:        depth + 1,
		})
	}

	return nil
}
