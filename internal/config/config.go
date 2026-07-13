package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type AppConfig struct {
	Verbose     bool
	ShowVersion bool
	Args        []string

	PrintAll bool

	RecentCommitCount int
	ParentScanDepth   int
	ParentScanDotDirs bool
	ParentScanHomeDir bool

	ProjectScanOutDirs  bool
	ProjectScanDepsDirs bool
	ProjectScanDepth    int

	ShowGoToolDeps bool

	DisableStructureOverview    bool
	DisableGitOverview          bool
	DisableWorkspaceOverview    bool
	DisableDependenciesOverview bool
	DisableContainerOverview    bool
	DisableToolsOverview        bool
	EnableCliOverview           bool
	EnableBehaviorInstruction   bool
	EnableOpencodeOverview      bool
}

func defaultAppConfig() *AppConfig {
	return &AppConfig{
		Verbose:     false,
		ShowVersion: false,
		Args:        []string{},

		PrintAll: false,

		RecentCommitCount: 12,
		ParentScanDepth:   2,
		ParentScanDotDirs: false,
		ParentScanHomeDir: false,

		ProjectScanOutDirs:  false,
		ProjectScanDepsDirs: false,
		ProjectScanDepth:    6,

		ShowGoToolDeps: true,

		DisableStructureOverview:    false,
		DisableGitOverview:          false,
		DisableWorkspaceOverview:    false,
		DisableDependenciesOverview: false,
		DisableContainerOverview:    false,
		DisableToolsOverview:        false,
		EnableCliOverview:           false,
		EnableBehaviorInstruction:   false,
		EnableOpencodeOverview:      false,
	}
}

func versionCommand(appConfig *AppConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Prints version message",
		Run: func(cmd *cobra.Command, args []string) {
			appConfig.Args = args
			appConfig.ShowVersion = true
		},
	}

	return cmd
}

func printCommand(appConfig *AppConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "print [projectRootPath]",
		Short: "Prints the raw exploration json",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			appConfig.Args = args
			appConfig.PrintAll = true
		},
	}

	return cmd
}

func loadEnvVars(appConfig *AppConfig) {
	EnvIsBool("VERBOSE", func(value bool) {
		appConfig.Verbose = value
	})

	EnvIsInt("RECENT_COMMIT_COUNT", func(value int) {
		appConfig.RecentCommitCount = value
	})
	EnvIsInt("PARENT_SCAN_DEPTH", func(value int) {
		appConfig.ParentScanDepth = value
	})
	EnvIsInt("PROJECT_SCAN_DEPTH", func(value int) {
		appConfig.ProjectScanDepth = value
	})
	EnvIsBool("PARENT_SCAN_DOT_DIRS", func(value bool) {
		appConfig.ParentScanDotDirs = value
	})
	EnvIsBool("PARENT_SCAN_HOME_DIR", func(value bool) {
		appConfig.ParentScanHomeDir = value
	})
	EnvIsBool("PROJECT_SCAN_OUT_DIRS", func(value bool) {
		appConfig.ProjectScanOutDirs = value
	})
	EnvIsBool("PROJECT_SCAN_DEPS_DIRS", func(value bool) {
		appConfig.ProjectScanDepsDirs = value
	})
	EnvIsBool("SHOW_GO_TOOL_DEPS", func(value bool) {
		appConfig.ShowGoToolDeps = value
	})
	EnvIsBool("DISABLE_STRUCTURE_OVERVIEW", func(value bool) {
		appConfig.DisableStructureOverview = value
	})
	EnvIsBool("DISABLE_GIT_OVERVIEW", func(value bool) {
		appConfig.DisableGitOverview = value
	})
	EnvIsBool("DISABLE_WORKSPACE_OVERVIEW", func(value bool) {
		appConfig.DisableWorkspaceOverview = value
	})
	EnvIsBool("DISABLE_DEPENDENCIES_OVERVIEW", func(value bool) {
		appConfig.DisableDependenciesOverview = value
	})
	EnvIsBool("DISABLE_CONTAINER_OVERVIEW", func(value bool) {
		appConfig.DisableContainerOverview = value
	})
	EnvIsBool("DISABLE_TOOLS_OVERVIEW", func(value bool) {
		appConfig.DisableToolsOverview = value
	})
	EnvIsBool("ENABLE_CLI_OVERVIEW", func(value bool) {
		appConfig.EnableCliOverview = value
	})
	EnvIsBool("ENABLE_BEHAVIOR_INSTRUCTION", func(value bool) {
		appConfig.EnableBehaviorInstruction = value
	})
	EnvIsBool("ENABLE_OPENCODE_OVERVIEW", func(value bool) {
		appConfig.EnableOpencodeOverview = value
	})
}

func applyExploreFlags(appConfig *AppConfig, cmd *cobra.Command) {
	// use uppercase shorthands for boolean (none-value) flags
	cmd.PersistentFlags().IntVarP(&appConfig.RecentCommitCount, "recent-commit-count", "c", appConfig.RecentCommitCount, "number of recent git commits to include (RECENT_COMMIT_COUNT)")
	cmd.PersistentFlags().IntVarP(&appConfig.ParentScanDepth, "parent-scan-depth", "p", appConfig.ParentScanDepth, "parent directory scan depth (PARENT_SCAN_DEPTH)")
	cmd.PersistentFlags().BoolVarP(&appConfig.ParentScanDotDirs, "parent-scan-dot-dirs", "D", appConfig.ParentScanDotDirs, "include dot directories during parent scan (PARENT_SCAN_DOT_DIRS)")
	cmd.PersistentFlags().BoolVarP(&appConfig.ParentScanHomeDir, "parent-scan-home-dir", "H", appConfig.ParentScanHomeDir, "include home directory during parent scan (PARENT_SCAN_HOME_DIR)")
	cmd.PersistentFlags().IntVarP(&appConfig.ProjectScanDepth, "project-scan-depth", "d", appConfig.ProjectScanDepth, "project structure scan depth (PROJECT_SCAN_DEPTH)")
	cmd.PersistentFlags().BoolVarP(&appConfig.ProjectScanOutDirs, "project-scan-out-dirs", "U", appConfig.ProjectScanOutDirs, "include dist, out and output directories in structure scan (PROJECT_SCAN_OUT_DIRS)")
	cmd.PersistentFlags().BoolVarP(&appConfig.ProjectScanDepsDirs, "project-scan-deps-dirs", "J", appConfig.ProjectScanDepsDirs, "include node_modules and vendor directories in structure scan (PROJECT_SCAN_DEPS_DIRS)")
	cmd.PersistentFlags().BoolVarP(&appConfig.ShowGoToolDeps, "show-go-tool-deps", "K", appConfig.ShowGoToolDeps, "label go.mod tool dependencies separately in dependencies overview (SHOW_GO_TOOL_DEPS)")
	cmd.PersistentFlags().BoolVarP(&appConfig.DisableStructureOverview, "disable-structure", "S", appConfig.DisableStructureOverview, "omit structure overview (DISABLE_STRUCTURE_OVERVIEW)")
	cmd.PersistentFlags().BoolVarP(&appConfig.DisableGitOverview, "disable-git", "G", appConfig.DisableGitOverview, "omit git overview (DISABLE_GIT_OVERVIEW)")
	cmd.PersistentFlags().BoolVarP(&appConfig.DisableWorkspaceOverview, "disable-workspace", "W", appConfig.DisableWorkspaceOverview, "omit workspace overview (DISABLE_WORKSPACE_OVERVIEW)")
	cmd.PersistentFlags().BoolVarP(&appConfig.DisableDependenciesOverview, "disable-dependencies", "E", appConfig.DisableDependenciesOverview, "omit dependencies overview (DISABLE_DEPENDENCIES_OVERVIEW)")
	cmd.PersistentFlags().BoolVarP(&appConfig.DisableContainerOverview, "disable-container", "C", appConfig.DisableContainerOverview, "omit container overview (DISABLE_CONTAINER_OVERVIEW)")
	cmd.PersistentFlags().BoolVarP(&appConfig.DisableToolsOverview, "disable-tools", "T", appConfig.DisableToolsOverview, "omit tools overview (DISABLE_TOOLS_OVERVIEW)")
	cmd.PersistentFlags().BoolVarP(&appConfig.EnableCliOverview, "enable-cli", "L", appConfig.EnableCliOverview, "include cli overview (ENABLE_CLI_OVERVIEW)")
	cmd.PersistentFlags().BoolVarP(&appConfig.EnableBehaviorInstruction, "enable-behavior", "B", appConfig.EnableBehaviorInstruction, "include behavior instructions (ENABLE_BEHAVIOR_INSTRUCTION)")
	cmd.PersistentFlags().BoolVarP(&appConfig.EnableOpencodeOverview, "enable-opencode", "O", appConfig.EnableOpencodeOverview, "include opencode overview (ENABLE_OPENCODE_OVERVIEW)")
}

func ParseConfig(
	displayName string,
	shortName string,
	version string,
	commit string,
) *AppConfig {
	appConfig := defaultAppConfig()

	rootCmd := &cobra.Command{
		Use: shortName,
		Short: displayName + " is a read-only MCP server for fast project context exploration.\n" +
			"For more help, visit https://github.com/NobleMajo/explorer-mcp",
		Run: func(cmd *cobra.Command, args []string) {
			appConfig.Args = args
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&appConfig.Verbose, "verbose", "b", appConfig.Verbose, "enable verbose mode (VERBOSE)")
	rootCmd.Flags().BoolVarP(&appConfig.ShowVersion, "version", "v", appConfig.ShowVersion, "prints version")

	applyExploreFlags(appConfig, rootCmd)

	loadEnvVars(appConfig)

	rootCmd.AddCommand(
		versionCommand(appConfig),
		printCommand(appConfig),
	)

	// wanted behavior: shows an error when using the "help" subcommand and does not execute
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "",
		Hidden: true,
	})

	cmd, err := rootCmd.ExecuteC()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if commandHelpRequested(cmd) {
		os.Exit(0)
	}

	if appConfig.Verbose {
		fmt.Fprintln(os.Stderr, "Verbose mode enabled")
	}

	if appConfig.ShowVersion {
		fmt.Println(displayName + " version " + version + ", build " + commit)
		os.Exit(0)
	}

	return appConfig
}

func commandHelpRequested(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if helpFlag := c.Flags().Lookup("help"); helpFlag != nil && helpFlag.Changed {
			return true
		}
	}
	return false
}
