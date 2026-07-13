package service

import (
	"github.com/NobleMajo/explorer-mcp/internal/config"
	"github.com/NobleMajo/explorer-mcp/internal/service/overviews/parent"
)

type exploreSettings struct {
	verbose                   bool
	recentCommitCount         int
	parentScanDepth           int
	parentScanDotDirs         bool
	parentScanHomeDir         bool
	repoScanDepth             int
	removeBehaviorInstruction bool
}

func exploreSettingsFromConfig(cfg *config.AppConfig) exploreSettings {
	if cfg == nil {
		return exploreSettings{}
	}

	return exploreSettings{
		verbose:                   cfg.Verbose,
		recentCommitCount:         cfg.RecentCommitCount,
		parentScanDepth:           cfg.ParentScanDepth,
		parentScanDotDirs:         cfg.ParentScanDotDirs,
		parentScanHomeDir:         cfg.ParentScanHomeDir,
		repoScanDepth:             cfg.RepoScanDepth,
		removeBehaviorInstruction: cfg.RemoveBehaviorInstruction,
	}
}

func (s exploreSettings) parentScanSettings() parent.ScanSettings {
	return parent.ScanSettings{
		Depth:       s.parentScanDepth,
		ScanDotDirs: s.parentScanDotDirs,
		ScanHomeDir: s.parentScanHomeDir,
	}
}
