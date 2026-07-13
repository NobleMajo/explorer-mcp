package parent

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
)

type ScanSettings struct {
	Depth         int
	ScanDotDirs   bool
	ScanHomeDir   bool
	HomeDir       string
}

type scanContext struct {
	settings ScanSettings
	homeDir  string
}

func listParentProjects(startDir string, settings ScanSettings) ([]string, error) {
	siblings := make([]string, 0)
	if settings.Depth < 1 {
		return siblings, nil
	}

	ctx, err := newScanContext(settings)
	if err != nil {
		return nil, err
	}

	cwd := filepath.Clean(startDir)
	seen := make(map[string]struct{})

	err = scanParentRepositories(cwd, ctx, func(absPath string, subfiles, subdirs []string) {
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

		siblings = append(siblings, formatSiblingProject(absPath, rel, subfiles, subdirs))
	})
	if err != nil {
		return nil, err
	}

	return siblings, nil
}

func newScanContext(settings ScanSettings) (scanContext, error) {
	ctx := scanContext{settings: settings}
	if settings.ScanHomeDir {
		return ctx, nil
	}

	homeDir := strings.TrimSpace(settings.HomeDir)
	if homeDir == "" {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return scanContext{}, err
		}
	}
	ctx.homeDir = filepath.Clean(homeDir)
	return ctx, nil
}

func scanParentRepositories(startDir string, ctx scanContext, callback func(path string, subfiles, subdirs []string)) error {
	current := filepath.Clean(startDir)
	previous := ""

	for p := 1; p <= ctx.settings.Depth; p++ {
		parent := filepath.Dir(current)
		if shouldStopParentScan(parent, current, ctx) {
			break
		}

		downDepth := ctx.settings.Depth - p + 1
		if err := scanDownwards(parent, previous, downDepth, 1, ctx, callback); err != nil {
			return err
		}

		previous = current
		current = parent
	}

	return nil
}

func shouldStopParentScan(parent, current string, ctx scanContext) bool {
	if parent == current {
		return true
	}
	if isFilesystemRoot(parent) {
		return true
	}
	if !ctx.settings.ScanHomeDir && ctx.homeDir != "" && filepath.Clean(parent) == ctx.homeDir {
		return true
	}
	return false
}

func isFilesystemRoot(path string) bool {
	path = filepath.Clean(path)
	if path == string(filepath.Separator) {
		return true
	}
	volume := filepath.VolumeName(path)
	if volume != "" && path == volume+string(filepath.Separator) {
		return true
	}
	return false
}

func scanDownwards(currentDir, skipDir string, maxDepth, currentDepth int, ctx scanContext, callback func(path string, subfiles, subdirs []string)) error {
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
		if !ctx.settings.ScanDotDirs && strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if globals.IsScanIgnored(entry.Name()) {
			continue
		}

		path := filepath.Clean(filepath.Join(currentDir, entry.Name()))
		if path == skipDir {
			continue
		}

		childEntries, err := os.ReadDir(path)
		if err != nil {
			continue
		}
		subfiles, subdirs := dirListingNames(childEntries)

		callback(path, subfiles, subdirs)

		if hasSiblingProjectFlags(path, subfiles, subdirs) {
			continue
		}

		if err := scanDownwards(path, "", maxDepth, currentDepth+1, ctx, callback); err != nil {
			return err
		}
	}

	return nil
}

func dirListingNames(entries []os.DirEntry) (subfiles, subdirs []string) {
	subfiles = make([]string, 0)
	subdirs = make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			subdirs = append(subdirs, entry.Name())
			continue
		}
		subfiles = append(subfiles, entry.Name())
	}
	sort.Strings(subfiles)
	sort.Strings(subdirs)
	return subfiles, subdirs
}

func hasNamedPathSegment(rel string) bool {
	for _, part := range strings.Split(rel, "/") {
		if part != ".." && part != "." && part != "" {
			return true
		}
	}
	return false
}
