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

	RecentCommitCount         int
	ParentScanDepth           int
	ParentScanDotDirs         bool
	ParentScanHomeDir         bool
	RepoScanDepth             int
	RemoveBehaviorInstruction bool
}

func defaultAppConfig() *AppConfig {
	return &AppConfig{
		Verbose:     false,
		ShowVersion: false,
		Args:        []string{},

		PrintAll: false,

		RecentCommitCount:         12,
		ParentScanDepth:           2,
		ParentScanDotDirs:         false,
		ParentScanHomeDir:         false,
		RepoScanDepth:             6,
		RemoveBehaviorInstruction: false,
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
		Use:   "print",
		Short: "Prints the raw exploration json",
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
	EnvIsInt("REPO_SCAN_DEPTH", func(value int) {
		appConfig.RepoScanDepth = value
	})
	EnvIsBool("PARENT_SCAN_DOT_DIRS", func(value bool) {
		appConfig.ParentScanDotDirs = value
	})
	EnvIsBool("PARENT_SCAN_HOME_DIR", func(value bool) {
		appConfig.ParentScanHomeDir = value
	})
	EnvIsBool("REMOVE_BEHAVIOR_INSTRUCTION", func(value bool) {
		appConfig.RemoveBehaviorInstruction = value
	})
}

func applyExploreFlags(appConfig *AppConfig, cmd *cobra.Command) {
	// use uppercase shorthands for boolean (none-value) flags
	cmd.PersistentFlags().IntVarP(&appConfig.RecentCommitCount, "recent-commit-count", "c", appConfig.RecentCommitCount, "number of recent git commits to include (RECENT_COMMIT_COUNT)")
	cmd.PersistentFlags().IntVarP(&appConfig.ParentScanDepth, "parent-scan-depth", "p", appConfig.ParentScanDepth, "parent directory scan depth (PARENT_SCAN_DEPTH)")
	cmd.PersistentFlags().BoolVarP(&appConfig.ParentScanDotDirs, "parent-scan-dot-dirs", "D", appConfig.ParentScanDotDirs, "include dot directories during parent scan (PARENT_SCAN_DOT_DIRS)")
	cmd.PersistentFlags().BoolVarP(&appConfig.ParentScanHomeDir, "parent-scan-home-dir", "H", appConfig.ParentScanHomeDir, "include home directory during parent scan (PARENT_SCAN_HOME_DIR)")
	cmd.PersistentFlags().IntVarP(&appConfig.RepoScanDepth, "repo-scan-depth", "d", appConfig.RepoScanDepth, "repo structure scan depth (REPO_SCAN_DEPTH)")
	cmd.PersistentFlags().BoolVarP(&appConfig.RemoveBehaviorInstruction, "no-behavior", "N", appConfig.RemoveBehaviorInstruction, "dont adds behavior instructions")
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
