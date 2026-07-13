package parent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestHasGitMetadata(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if hasGitMetadata(root) {
		t.Fatal("expected no git metadata in empty dir")
	}

	gitDir := filepath.Join(root, ".git")
	if err := os.Mkdir(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if !hasGitMetadata(root) {
		t.Fatal("expected git metadata when .git exists")
	}
}

func TestListSiblingProjects(t *testing.T) {
	parent := t.TempDir()
	current := filepath.Join(parent, "alpha")
	sibling := filepath.Join(parent, "beta")
	if err := os.Mkdir(current, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(sibling, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(sibling, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, filepath.Join(parent, "notes.txt"), "ignore me\n")

	gitSiblings, siblings, err := listSiblingProjects(parent, current)
	if err != nil {
		t.Fatalf("listSiblingProjects() error: %v", err)
	}
	if len(gitSiblings) != 1 || len(siblings) != 0 {
		t.Fatalf("gitSiblings=%v siblings=%v, want [../beta] and []", gitSiblings, siblings)
	}
	if gitSiblings[0] != "../beta" {
		t.Fatalf("git sibling = %q, want ../beta", gitSiblings[0])
	}
}

func TestWorkspaceContext(t *testing.T) {
	parent := t.TempDir()
	current := filepath.Join(parent, "app")
	sibling := filepath.Join(parent, "other")
	if err := os.Mkdir(current, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(sibling, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.Chdir(t, current)

	result, err := ParentOverview()(false)
	if err != nil {
		t.Fatalf("ParentOverview() error: %v", err)
	}

	resp, ok := result.(workspaceContextResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	if resp.CurrentWorkingDirectoryPath != current || resp.ParentDirectoryPath != parent {
		t.Fatalf("unexpected paths: cwd=%q parent=%q", resp.CurrentWorkingDirectoryPath, resp.ParentDirectoryPath)
	}
	if resp.SiblingProjectCount != 1 || len(resp.GitSiblingProjects) != 0 || len(resp.SiblingProjects) != 1 {
		t.Fatalf("unexpected siblings: git=%v other=%v", resp.GitSiblingProjects, resp.SiblingProjects)
	}
	if resp.SiblingProjects[0] != "../other" {
		t.Fatalf("relativePath = %q, want ../other", resp.SiblingProjects[0])
	}
}

func TestWorkspaceContextUnreadableParent(t *testing.T) {
	parent := t.TempDir()
	current := filepath.Join(parent, "app")
	if err := os.Mkdir(current, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.Chdir(t, current)
	if err := os.Chmod(parent, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(parent, 0o755) })

	_, err := ParentOverview()(false)
	if err == nil {
		t.Fatal("expected error when parent directory is unreadable")
	}
}

func TestListSiblingProjectsMissingParent(t *testing.T) {
	_, _, err := listSiblingProjects("/does/not/exist", "/does/not/exist/child")
	if err == nil {
		t.Fatal("expected error for missing parent directory")
	}
}
