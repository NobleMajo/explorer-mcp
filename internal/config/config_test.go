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
}

func TestParseConfigPrintExploreFlags(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	os.Args = []string{
		"explorer-mcp", "print",
		"-c", "0",
		"-p", "1",
		"-d", "2",
		"-n",
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
	if !cfg.RemoveBehaviorInstruction {
		t.Fatal("expected RemoveBehaviorInstruction true")
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
