package service

import "github.com/NobleMajo/explorer-mcp/internal/config"

type exploreSettings struct {
	verbose                   bool
	recentCommitCount         int
	parentScanDepth           int
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
		repoScanDepth:             cfg.RepoScanDepth,
		removeBehaviorInstruction: cfg.RemoveBehaviorInstruction,
	}
}
