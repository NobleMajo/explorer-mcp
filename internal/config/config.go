package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type AppConfig struct {
	Verbose     bool
	ShowVersion bool
	ShowHelp    bool
	Args        []string

	PrintAll bool

	RecentCommitCount         int
	ParentScanDepth           int
	RepoScanDepth             int
	ScanIgnoreFile            string
	RemoveBehaviorInstruction bool
}

func defaultAppConfig() *AppConfig {
	return &AppConfig{
		Verbose:     false,
		ShowVersion: false,
		ShowHelp:    false,
		Args:        []string{},

		PrintAll: false,

		RecentCommitCount:         10,
		ParentScanDepth:           3,
		RepoScanDepth:             7,
		ScanIgnoreFile:            ".gitignore",
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
	EnvIsString("SCAN_IGNORE_FILE", func(value string) {
		appConfig.ScanIgnoreFile = value
	})
	EnvIsBool("ADD_BEHAVIOR_INSTRUCTION", func(value bool) {
		appConfig.RemoveBehaviorInstruction = value
	})

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
		Run: func(cmd *cobra.Command, args []string) {},
	}

	rootCmd.PersistentFlags().BoolVarP(&appConfig.Verbose, "verbose", "b", appConfig.Verbose, "enable verbose mode (VERBOSE)")
	rootCmd.Flags().BoolVarP(&appConfig.ShowVersion, "version", "v", appConfig.ShowVersion, "prints version")

	rootCmd.Flags().IntVarP(&appConfig.RecentCommitCount, "recent-commit-count", "c", appConfig.RecentCommitCount, "number of recent git commits to include (RECENT_COMMIT_COUNT)")
	rootCmd.Flags().IntVarP(&appConfig.ParentScanDepth, "parent-scan-depth", "p", appConfig.ParentScanDepth, "parent directory scan depth (PARENT_SCAN_DEPTH)")
	rootCmd.Flags().IntVarP(&appConfig.RepoScanDepth, "repo-scan-depth", "d", appConfig.RepoScanDepth, "repo structure scan depth (REPO_SCAN_DEPTH)")

	rootCmd.Flags().StringVarP(&appConfig.ScanIgnoreFile, "ignore-file", "i", appConfig.ScanIgnoreFile, "overwrites default .gitignore file for scans")
	rootCmd.Flags().BoolVarP(&appConfig.RemoveBehaviorInstruction, "no-behavior", "n", appConfig.RemoveBehaviorInstruction, "dont adds behavior instructions")

	loadEnvVars(appConfig)

	rootCmd.AddCommand(
		versionCommand(appConfig),
		printCommand(appConfig),
	)

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if appConfig.Verbose {
		fmt.Fprintln(os.Stderr, "Verbose mode enabled")
	}

	if appConfig.ShowVersion {
		fmt.Println(displayName + " version " + version + ", build " + commit)
		os.Exit(0)
	}

	if appConfig.ShowHelp {
		rootCmd.Help()
		os.Exit(0)
	}

	return appConfig
}
