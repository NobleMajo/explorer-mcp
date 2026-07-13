package service

import (
	"errors"

	"github.com/NobleMajo/explorer-mcp/internal/config"
	"github.com/NobleMajo/explorer-mcp/internal/service/overviews/parent"
)

var ErrAllOverviewsDisabled = errors.New("all overviews are disabled; enable at least one overview section")

type exploreSettings struct {
	verbose                     bool
	recentCommitCount           int
	parentScanDepth             int
	parentScanDotDirs           bool
	parentScanHomeDir           bool
	repoScanDepth               int
	enableBehaviorInstruction   bool
	disableStructureOverview    bool
	disableGitOverview          bool
	disableWorkspaceOverview    bool
	disableDependenciesOverview bool
	disableContainerOverview    bool
	disableToolsOverview        bool
	enableCliOverview           bool
	enableOpencodeOverview      bool
}

func exploreSettingsFromConfig(cfg *config.AppConfig) exploreSettings {
	if cfg == nil {
		return exploreSettings{}
	}

	return exploreSettings{
		verbose:                     cfg.Verbose,
		recentCommitCount:           cfg.RecentCommitCount,
		parentScanDepth:             cfg.ParentScanDepth,
		parentScanDotDirs:           cfg.ParentScanDotDirs,
		parentScanHomeDir:           cfg.ParentScanHomeDir,
		repoScanDepth:               cfg.RepoScanDepth,
		enableBehaviorInstruction:   cfg.EnableBehaviorInstruction,
		disableStructureOverview:    cfg.DisableStructureOverview,
		disableGitOverview:          cfg.DisableGitOverview,
		disableWorkspaceOverview:    cfg.DisableWorkspaceOverview,
		disableDependenciesOverview: cfg.DisableDependenciesOverview,
		disableContainerOverview:    cfg.DisableContainerOverview,
		disableToolsOverview:        cfg.DisableToolsOverview,
		enableCliOverview:           cfg.EnableCliOverview,
		enableOpencodeOverview:      cfg.EnableOpencodeOverview,
	}
}

func (s exploreSettings) parentScanSettings() parent.ScanSettings {
	return parent.ScanSettings{
		Depth:       s.parentScanDepth,
		ScanDotDirs: s.parentScanDotDirs,
		ScanHomeDir: s.parentScanHomeDir,
	}
}

func (s exploreSettings) hasEnabledOverview() bool {
	return !s.disableStructureOverview ||
		!s.disableGitOverview ||
		!s.disableWorkspaceOverview ||
		!s.disableDependenciesOverview ||
		!s.disableContainerOverview ||
		!s.disableToolsOverview ||
		s.enableCliOverview ||
		s.enableOpencodeOverview
}
