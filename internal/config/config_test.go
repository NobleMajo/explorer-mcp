package config

import (
	"os"
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
