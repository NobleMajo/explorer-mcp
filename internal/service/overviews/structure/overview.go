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
	RootPath              string           `json:"rootPath"`
	MaxDepth              int              `json:"maxDepth"`
	FollowGitIgnore       bool             `json:"followGitIgnore"`
	IgnoredDirectoryNames []string         `json:"ignoredDirectoryNames"`
	IgnoredFileNames      []string         `json:"ignoredFileNames"`
	EntryCount            int              `json:"entryCount"`
	Entries               []structureEntry `json:"entries"`
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

	ignoredDirNames := append([]string(nil), globals.ScanIgnoreFiles...)
	sort.Strings(ignoredDirNames)
	ignoredFileNames := append([]string(nil), globals.IgnoreFiles...)
	sort.Strings(ignoredFileNames)

	return repoStructureResponse{
		Meta: jsonresp.Meta{
			ToolName:      "repo_structure",
			SchemaVersion: jsonresp.SchemaVersion,
		},
		RootPath:              root,
		MaxDepth:              globals.StructureScanMaxDepth,
		FollowGitIgnore:       globals.FollowGitIgnore,
		IgnoredDirectoryNames: ignoredDirNames,
		IgnoredFileNames:      ignoredFileNames,
		EntryCount:            len(entries),
		Entries:               entries,
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
		if dirEntries[i].IsDir() != dirEntries[j].IsDir() {
			return dirEntries[i].IsDir()
		}
		return dirEntries[i].Name() < dirEntries[j].Name()
	})

	for _, entry := range dirEntries {
		if globals.IsScanIgnored(entry.Name()) {
			continue
		}
		if !entry.IsDir() && globals.IsIgnoredFile(entry.Name()) {
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

		relPath = filepath.ToSlash(relPath)
		if entry.IsDir() {
			relPath += "/"
		}

		*entries = append(*entries, structureEntry{
			RelativePath: relPath,
			EntryName:    entry.Name(),
			IsDirectory:  entry.IsDir(),
			Depth:        depth + 1,
		})

		if entry.IsDir() {
			if err := appendStructureEntries(root, fullPath, depth+1, entries, state); err != nil {
				return err
			}
		}
	}

	return nil
}
