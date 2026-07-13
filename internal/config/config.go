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
	DirectOut   bool
	Args        []string
}

func defaultAppConfig() *AppConfig {
	return &AppConfig{
		Verbose:     false,
		ShowVersion: false,
		ShowHelp:    false,
		Args:        []string{},
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

func loadEnvVars(appConfig *AppConfig) {
	EnvIsBool("VERBOSE", func(value bool) {
		appConfig.Verbose = value
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
		Short: displayName + " is a read-only MCP server for Git repository exploration.\n" +
			"For more help, visit https://github.com/NobleMajo/explorer-mcp",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	rootCmd.PersistentFlags().BoolVarP(&appConfig.Verbose, "verbose", "b", appConfig.Verbose, "enable verbose mode (VERBOSE)")
	rootCmd.Flags().BoolVarP(&appConfig.ShowVersion, "version", "v", appConfig.ShowVersion, "prints version")
	rootCmd.Flags().BoolVar(&appConfig.DirectOut, "out", appConfig.DirectOut, "print explore JSON to stdout and exit")

	loadEnvVars(appConfig)

	rootCmd.AddCommand(
		versionCommand(appConfig),
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
