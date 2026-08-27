package agentc

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestAgentcOverviewEmpty(t *testing.T) {
	root := t.TempDir()

	result, err := AgentcOverview()(root, false)
	if err != nil {
		t.Fatalf("AgentcOverview() error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil when no agent files, got %v", result)
	}
}

func TestAgentcOverviewFindsRootFiles(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "AGENTS.md"), "# agents\n")
	testutil.WriteFile(t, filepath.Join(root, "CLAUDE.md"), "# claude\n")
	testutil.WriteFile(t, filepath.Join(root, "CONTRIBUTING.md"), "# contrib\n")
	// content must not matter / must not be returned
	testutil.WriteFile(t, filepath.Join(root, "README.md"), "# readme ignored\n")

	result, err := AgentcOverview()(root, false)
	if err != nil {
		t.Fatalf("AgentcOverview() error: %v", err)
	}

	got, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected type %T", result)
	}

	want := []string{"AGENTS.md", "CLAUDE.md", "CONTRIBUTING.md"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAgentcOverviewCursorRulesAndDocs(t *testing.T) {
	root := t.TempDir()
	rulesDir := filepath.Join(root, ".cursor", "rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, filepath.Join(rulesDir, "general.mdc"), "rule body\n")
	testutil.WriteFile(t, filepath.Join(rulesDir, "extra.md"), "rule body\n")
	testutil.WriteFile(t, filepath.Join(rulesDir, ".DS_Store"), "noise\n")
	if err := os.Mkdir(filepath.Join(rulesDir, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, filepath.Join(rulesDir, "nested", "skip.md"), "ignored nested\n")

	docsDir := filepath.Join(root, "docs")
	if err := os.Mkdir(docsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, filepath.Join(docsDir, "guide.md"), "docs\n")
	testutil.WriteFile(t, filepath.Join(docsDir, "notes.mdx"), "docs\n")
	testutil.WriteFile(t, filepath.Join(docsDir, ".hidden.md"), "ignored hidden\n")
	testutil.WriteFile(t, filepath.Join(docsDir, "image.png"), "nope\n")
	if err := os.Mkdir(filepath.Join(docsDir, "deep"), 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, filepath.Join(docsDir, "deep", "skip.md"), "ignored deep\n")

	result, err := AgentcOverview()(root, false)
	if err != nil {
		t.Fatalf("AgentcOverview() error: %v", err)
	}

	got, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected type %T", result)
	}

	want := []string{
		".cursor/rules/extra.md",
		".cursor/rules/general.mdc",
		"docs/guide.md",
		"docs/notes.mdx",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAgentcOverviewCursorrulesAndAgentMd(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, ".cursorrules"), "legacy\n")
	testutil.WriteFile(t, filepath.Join(root, "AGENT.md"), "agent\n")

	result, err := AgentcOverview()(root, false)
	if err != nil {
		t.Fatalf("AgentcOverview() error: %v", err)
	}

	got, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected type %T", result)
	}

	want := []string{".cursorrules", "AGENT.md"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAgentcOverviewIgnoresDirectoryNamedLikeRootFile(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "AGENTS.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, filepath.Join(root, "CLAUDE.md"), "# claude\n")

	result, err := AgentcOverview()(root, false)
	if err != nil {
		t.Fatalf("AgentcOverview() error: %v", err)
	}

	got, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected type %T", result)
	}
	want := []string{"CLAUDE.md"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAgentcOverviewUnreadableDocsDirStillReturnsRoots(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "AGENTS.md"), "# agents\n")
	docsDir := filepath.Join(root, "docs")
	if err := os.Mkdir(docsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, filepath.Join(docsDir, "guide.md"), "docs\n")
	if err := os.Chmod(docsDir, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(docsDir, 0o755) })

	result, err := AgentcOverview()(root, false)
	if err != nil {
		t.Fatalf("AgentcOverview() error: %v", err)
	}

	got, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected type %T", result)
	}
	want := []string{"AGENTS.md"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v (unreadable docs should be skipped)", got, want)
	}
}
