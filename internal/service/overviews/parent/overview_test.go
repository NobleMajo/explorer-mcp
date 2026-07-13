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

	got, err := listSiblingProjects(parent, current)
	if err != nil {
		t.Fatalf("listSiblingProjects() error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(siblings) = %d, want 2", len(got))
	}

	var currentProject, betaProject *siblingProject
	for i := range got {
		switch got[i].DirectoryName {
		case "alpha":
			currentProject = &got[i]
		case "beta":
			betaProject = &got[i]
		}
	}
	if currentProject == nil || !currentProject.IsCurrentProject || currentProject.IsGitRepo {
		t.Fatalf("unexpected alpha sibling: %+v", currentProject)
	}
	if betaProject == nil || betaProject.IsCurrentProject || !betaProject.IsGitRepo {
		t.Fatalf("unexpected beta sibling: %+v", betaProject)
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

	if resp.ToolName != "workspace_context" {
		t.Fatalf("toolName = %q", resp.ToolName)
	}
	if resp.CurrentWorkingDirectoryPath != current || resp.ParentDirectoryPath != parent {
		t.Fatalf("unexpected paths: cwd=%q parent=%q", resp.CurrentWorkingDirectoryPath, resp.ParentDirectoryPath)
	}
	if resp.SiblingProjectCount != 2 || len(resp.SiblingProjects) != 2 {
		t.Fatalf("unexpected siblings: %+v", resp.SiblingProjects)
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
	_, err := listSiblingProjects("/does/not/exist", "/does/not/exist/child")
	if err == nil {
		t.Fatal("expected error for missing parent directory")
	}
}
