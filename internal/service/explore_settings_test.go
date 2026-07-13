package service

import (
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/config"
)

func TestHasEnabledOverview(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    exploreSettings
		want bool
	}{
		{
			name: "all enabled",
			s:    exploreSettings{},
			want: true,
		},
		{
			name: "all disabled",
			s: exploreSettings{
				disableStructureOverview:    true,
				disableGitOverview:          true,
				disableWorkspaceOverview:    true,
				disableDependenciesOverview: true,
				disableContainerOverview:    true,
				disableToolsOverview:        true,
			},
			want: false,
		},
		{
			name: "only cli enabled",
			s: exploreSettings{
				disableStructureOverview:    true,
				disableGitOverview:          true,
				disableWorkspaceOverview:    true,
				disableDependenciesOverview: true,
				disableContainerOverview:    true,
				disableToolsOverview:        true,
				enableCliOverview:           true,
			},
			want: true,
		},
		{
			name: "only structure enabled",
			s: exploreSettings{
				disableGitOverview:          true,
				disableWorkspaceOverview:    true,
				disableDependenciesOverview: true,
				disableContainerOverview:    true,
				disableToolsOverview:        true,
			},
			want: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.s.hasEnabledOverview(); got != tc.want {
				t.Fatalf("hasEnabledOverview() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestExploreSettingsFromConfigMapsFields(t *testing.T) {
	cfg := &config.AppConfig{
		Verbose:                     true,
		RecentCommitCount:           5,
		ParentScanDepth:             4,
		ParentScanDotDirs:           true,
		ParentScanHomeDir:           true,
		RepoScanDepth:               3,
		DisableBehaviorInstruction:  true,
		DisableStructureOverview:    true,
		DisableGitOverview:          true,
		DisableWorkspaceOverview:    true,
		DisableDependenciesOverview: true,
		DisableContainerOverview:    true,
		DisableToolsOverview:        true,
		EnableCliOverview:           true,
	}

	settings := exploreSettingsFromConfig(cfg)
	if settings.verbose != cfg.Verbose ||
		settings.recentCommitCount != cfg.RecentCommitCount ||
		settings.parentScanDepth != cfg.ParentScanDepth ||
		settings.parentScanDotDirs != cfg.ParentScanDotDirs ||
		settings.parentScanHomeDir != cfg.ParentScanHomeDir ||
		settings.repoScanDepth != cfg.RepoScanDepth ||
		settings.disableBehaviorInstruction != cfg.DisableBehaviorInstruction ||
		settings.disableStructureOverview != cfg.DisableStructureOverview ||
		settings.disableGitOverview != cfg.DisableGitOverview ||
		settings.disableWorkspaceOverview != cfg.DisableWorkspaceOverview ||
		settings.disableDependenciesOverview != cfg.DisableDependenciesOverview ||
		settings.disableContainerOverview != cfg.DisableContainerOverview ||
		settings.disableToolsOverview != cfg.DisableToolsOverview ||
		settings.enableCliOverview != cfg.EnableCliOverview {
		t.Fatalf("exploreSettingsFromConfig() = %+v, want mapped fields from cfg", settings)
	}

	if settings.parentScanSettings().Depth != cfg.ParentScanDepth {
		t.Fatalf("parentScanSettings().Depth = %d, want %d", settings.parentScanSettings().Depth, cfg.ParentScanDepth)
	}
}

func TestExploreSettingsFromConfigNil(t *testing.T) {
	settings := exploreSettingsFromConfig(nil)
	if !settings.hasEnabledOverview() {
		t.Fatal("expected nil config to keep default overview sections enabled")
	}
	if settings.enableCliOverview {
		t.Fatal("expected cli disabled by default")
	}
}
