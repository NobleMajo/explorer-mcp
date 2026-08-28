package parent

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/NobleMajo/explorer-mcp/internal/service/workspacescan"
)

type ScanSettings struct {
	Depth       int
	ScanDotDirs bool
	ScanHomeDir bool
	HomeDir     string
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

	err = scanParentRepositories(cwd, ctx, func(listing workspacescan.Listing) {
		rel, relErr := filepath.Rel(cwd, listing.AbsPath)
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

		siblings = append(siblings, workspacescan.FormatPathWithFlags(rel, listing.Flags))
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

func scanParentRepositories(startDir string, ctx scanContext, callback func(listing workspacescan.Listing)) error {
	current := filepath.Clean(startDir)
	previous := ""

	for p := 1; p <= ctx.settings.Depth; p++ {
		parent := filepath.Dir(current)
		if shouldStopParentScan(parent, current, ctx) {
			break
		}

		downDepth := ctx.settings.Depth - p + 1
		if err := scanDownwards(parent, previous, downDepth, ctx, callback); err != nil {
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

func scanDownwards(currentDir, skipDir string, maxDepth int, ctx scanContext, callback func(listing workspacescan.Listing)) error {
	opts := workspacescan.SiblingWalkOptions(currentDir, maxDepth, ctx.settings.ScanDotDirs, skipDir)

	return workspacescan.WalkDown(currentDir, opts, func(listing workspacescan.Listing) error {
		callback(listing)
		return nil
	})
}

func hasNamedPathSegment(rel string) bool {
	for _, part := range strings.Split(rel, "/") {
		if part != ".." && part != "." && part != "" {
			return true
		}
	}
	return false
}
