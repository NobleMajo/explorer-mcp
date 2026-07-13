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
			name: "only opencode enabled",
			s: exploreSettings{
				disableStructureOverview:    true,
				disableGitOverview:          true,
				disableWorkspaceOverview:    true,
				disableDependenciesOverview: true,
				disableContainerOverview:    true,
				disableToolsOverview:        true,
				enableOpencodeOverview:      true,
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
		ProjectScanDepth:            3,
		ProjectScanOutDirs:          true,
		ProjectScanDepsDirs:         true,
		EnableBehaviorInstruction:   true,
		DisableStructureOverview:    true,
		DisableGitOverview:          true,
		DisableWorkspaceOverview:    true,
		DisableDependenciesOverview: true,
		DisableContainerOverview:    true,
		DisableToolsOverview:        true,
		EnableCliOverview:           true,
		EnableOpencodeOverview:      true,
	}

	settings := exploreSettingsFromConfig(cfg)
	if settings.verbose != cfg.Verbose ||
		settings.recentCommitCount != cfg.RecentCommitCount ||
		settings.parentScanDepth != cfg.ParentScanDepth ||
		settings.parentScanDotDirs != cfg.ParentScanDotDirs ||
		settings.parentScanHomeDir != cfg.ParentScanHomeDir ||
		settings.projectScanDepth != cfg.ProjectScanDepth ||
		settings.projectScanOutDirs != cfg.ProjectScanOutDirs ||
		settings.projectScanDepsDirs != cfg.ProjectScanDepsDirs ||
		settings.enableBehaviorInstruction != cfg.EnableBehaviorInstruction ||
		settings.disableStructureOverview != cfg.DisableStructureOverview ||
		settings.disableGitOverview != cfg.DisableGitOverview ||
		settings.disableWorkspaceOverview != cfg.DisableWorkspaceOverview ||
		settings.disableDependenciesOverview != cfg.DisableDependenciesOverview ||
		settings.disableContainerOverview != cfg.DisableContainerOverview ||
		settings.disableToolsOverview != cfg.DisableToolsOverview ||
		settings.enableCliOverview != cfg.EnableCliOverview ||
		settings.enableOpencodeOverview != cfg.EnableOpencodeOverview {
		t.Fatalf("exploreSettingsFromConfig() = %+v, want mapped fields from cfg", settings)
	}

	if settings.parentScanSettings().Depth != cfg.ParentScanDepth {
		t.Fatalf("parentScanSettings().Depth = %d, want %d", settings.parentScanSettings().Depth, cfg.ParentScanDepth)
	}
	if settings.projectScanSettings().Depth != cfg.ProjectScanDepth {
		t.Fatalf("projectScanSettings().Depth = %d, want %d", settings.projectScanSettings().Depth, cfg.ProjectScanDepth)
	}
}

func TestProjectScanSettingsMapsCollapseFlags(t *testing.T) {
	settings := exploreSettings{
		projectScanDepth:    5,
		projectScanOutDirs:  true,
		projectScanDepsDirs: true,
	}

	scan := settings.projectScanSettings()
	if scan.Depth != 5 {
		t.Fatalf("Depth = %d, want 5", scan.Depth)
	}
	if !scan.OutDirs {
		t.Fatal("expected OutDirs true")
	}
	if !scan.DepsDirs {
		t.Fatal("expected DepsDirs true")
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
	if settings.enableOpencodeOverview {
		t.Fatal("expected opencode disabled by default")
	}
}
