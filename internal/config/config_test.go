package config

import (
	"os"
	"os/exec"
	"testing"
)

func TestParseConfigPrintCommand(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	os.Args = []string{"explorer-mcp", "print"}
	cfg := ParseConfig("Demo", "demo", "1.0.0", "abc")
	if !cfg.PrintAll {
		t.Fatal("expected PrintAll true")
	}
	if cfg.EnableCliOverview {
		t.Fatal("expected EnableCliOverview false by default")
	}
}

func TestParseConfigParentScanFlags(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	os.Args = []string{"explorer-mcp", "print", "-D", "-H"}
	cfg := ParseConfig("Demo", "demo", "1.0.0", "abc")

	if !cfg.PrintAll {
		t.Fatal("expected PrintAll true")
	}
	if !cfg.ParentScanDotDirs {
		t.Fatal("expected ParentScanDotDirs true")
	}
	if !cfg.ParentScanHomeDir {
		t.Fatal("expected ParentScanHomeDir true")
	}
}

func TestParseConfigPrintExploreFlags(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	os.Args = []string{
		"explorer-mcp", "print",
		"-c", "0",
		"-p", "1",
		"-d", "2",
		"-N",
	}
	cfg := ParseConfig("Demo", "demo", "1.0.0", "abc")

	if !cfg.PrintAll {
		t.Fatal("expected PrintAll true")
	}
	if cfg.RecentCommitCount != 0 {
		t.Fatalf("RecentCommitCount = %d, want 0", cfg.RecentCommitCount)
	}
	if cfg.ParentScanDepth != 1 {
		t.Fatalf("ParentScanDepth = %d, want 1", cfg.ParentScanDepth)
	}
	if cfg.RepoScanDepth != 2 {
		t.Fatalf("RepoScanDepth = %d, want 2", cfg.RepoScanDepth)
	}
	if !cfg.DisableBehaviorInstruction {
		t.Fatal("expected DisableBehaviorInstruction true")
	}
}

func TestParseConfigDisableOverviewFlags(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	os.Args = []string{
		"explorer-mcp", "print",
		"-S", "-G", "-W", "-E", "-C", "-T",
	}
	cfg := ParseConfig("Demo", "demo", "1.0.0", "abc")

	if !cfg.PrintAll {
		t.Fatal("expected PrintAll true")
	}
	if !cfg.DisableStructureOverview {
		t.Fatal("expected DisableStructureOverview true")
	}
	if !cfg.DisableGitOverview {
		t.Fatal("expected DisableGitOverview true")
	}
	if !cfg.DisableWorkspaceOverview {
		t.Fatal("expected DisableWorkspaceOverview true")
	}
	if !cfg.DisableDependenciesOverview {
		t.Fatal("expected DisableDependenciesOverview true")
	}
	if !cfg.DisableContainerOverview {
		t.Fatal("expected DisableContainerOverview true")
	}
	if !cfg.DisableToolsOverview {
		t.Fatal("expected DisableToolsOverview true")
	}
	if cfg.EnableCliOverview {
		t.Fatal("expected EnableCliOverview false by default")
	}
}

func TestParseConfigEnableCliOverviewFlag(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	os.Args = []string{"explorer-mcp", "print", "-L"}
	cfg := ParseConfig("Demo", "demo", "1.0.0", "abc")

	if !cfg.EnableCliOverview {
		t.Fatal("expected EnableCliOverview true")
	}
}

func TestParseConfigEnableCliOverviewEnv(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	t.Setenv("ENABLE_CLI_OVERVIEW", "true")
	os.Args = []string{"explorer-mcp", "print"}
	cfg := ParseConfig("Demo", "demo", "1.0.0", "abc")

	if !cfg.EnableCliOverview {
		t.Fatal("expected ENABLE_CLI_OVERVIEW env to enable cli overview")
	}
}

func TestParseConfigDisableBehaviorInstructionEnv(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	t.Setenv("DISABLE_BEHAVIOR_INSTRUCTION", "true")
	os.Args = []string{"explorer-mcp", "print"}
	cfg := ParseConfig("Demo", "demo", "1.0.0", "abc")

	if !cfg.DisableBehaviorInstruction {
		t.Fatal("expected DISABLE_BEHAVIOR_INSTRUCTION env to disable behavior")
	}
}

func TestParseConfigExploreFlagsBeforePrint(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	os.Args = []string{
		"explorer-mcp",
		"-c", "0",
		"-d", "0",
		"print",
	}
	cfg := ParseConfig("Demo", "demo", "1.0.0", "abc")

	if !cfg.PrintAll {
		t.Fatal("expected PrintAll true")
	}
	if cfg.RecentCommitCount != 0 {
		t.Fatalf("RecentCommitCount = %d, want 0", cfg.RecentCommitCount)
	}
	if cfg.RepoScanDepth != 0 {
		t.Fatalf("RepoScanDepth = %d, want 0", cfg.RepoScanDepth)
	}
}

func TestParseConfigRejectsUnknownSubcommand(t *testing.T) {
	if os.Getenv("TEST_PARSE_CONFIG_SUBCMD") == "1" {
		os.Args = []string{"explorer-mcp", "help"}
		ParseConfig("Demo", "demo", "1.0.0", "abc")
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestParseConfigRejectsUnknownSubcommand$")
	cmd.Env = append(os.Environ(), "TEST_PARSE_CONFIG_SUBCMD=1")
	err := cmd.Run()
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %v", err)
	}
	if exitErr.ExitCode() != 1 {
		t.Fatalf("exit code = %d, want 1", exitErr.ExitCode())
	}
}

func TestParseConfigPrintHelpExits(t *testing.T) {
	if os.Getenv("TEST_PARSE_CONFIG_PRINT_HELP") == "1" {
		os.Args = []string{"explorer-mcp", "print", "--help"}
		ParseConfig("Demo", "demo", "1.0.0", "abc")
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestParseConfigPrintHelpExits$")
	cmd.Env = append(os.Environ(), "TEST_PARSE_CONFIG_PRINT_HELP=1")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("expected exit 0, got %v", err)
	}
}

func TestParseConfigRootHelpExits(t *testing.T) {
	if os.Getenv("TEST_PARSE_CONFIG_ROOT_HELP") == "1" {
		os.Args = []string{"explorer-mcp", "--help"}
		ParseConfig("Demo", "demo", "1.0.0", "abc")
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestParseConfigRootHelpExits$")
	cmd.Env = append(os.Environ(), "TEST_PARSE_CONFIG_ROOT_HELP=1")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("expected exit 0, got %v", err)
	}
}
