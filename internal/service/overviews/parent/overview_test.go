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

	result, err := ParentOverview(1)()(false)
	if err != nil {
		t.Fatalf("ParentOverview() error: %v", err)
	}

	resp, ok := result.(workspaceContextResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	if resp.ParentScanPerformed != true {
		t.Fatalf("parentScanPerformed = %v, want true", resp.ParentScanPerformed)
	}
	if resp.SiblingProjectCount == nil || *resp.SiblingProjectCount != 1 || len(resp.SiblingProjects) != 1 {
		t.Fatalf("unexpected siblings: %v", resp.SiblingProjects)
	}
	if siblingRelativePath(resp.SiblingProjects[0]) != "../other" {
		t.Fatalf("relativePath = %q, want ../other", resp.SiblingProjects[0])
	}
}

func TestWorkspaceContextSkipsParentScanAtDepthZero(t *testing.T) {
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

	result, err := ParentOverview(0)()(false)
	if err != nil {
		t.Fatalf("ParentOverview() error: %v", err)
	}

	resp, ok := result.(workspaceContextResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if resp.ParentScanPerformed {
		t.Fatal("expected parentScanPerformed false")
	}
	if resp.SiblingProjectCount != nil || resp.SiblingProjects != nil {
		t.Fatalf("expected no sibling scan data, got count=%v siblings=%v", resp.SiblingProjectCount, resp.SiblingProjects)
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

	_, err := ParentOverview(1)()(false)
	if err == nil {
		t.Fatal("expected error when parent directory is unreadable")
	}
}
