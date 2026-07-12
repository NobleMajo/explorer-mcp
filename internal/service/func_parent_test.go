package service

import (
	"os"
	"path/filepath"
	"testing"
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
	writeFile(t, filepath.Join(parent, "notes.txt"), "ignore me\n")

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
	if err := os.Mkdir(current, 0o755); err != nil {
		t.Fatal(err)
	}
	chdir(t, current)

	jsonText, err := WorkspaceContext()
	if err != nil {
		t.Fatalf("WorkspaceContext() error: %v", err)
	}

	var resp workspaceContextResponse
	parseJSONResponse(t, jsonText, &resp)

	if resp.ToolName != "workspace_context" {
		t.Fatalf("toolName = %q", resp.ToolName)
	}
	if resp.CurrentWorkingDirectoryPath != current || resp.ParentDirectoryPath != parent {
		t.Fatalf("unexpected paths: cwd=%q parent=%q", resp.CurrentWorkingDirectoryPath, resp.ParentDirectoryPath)
	}
	if resp.SiblingProjectCount != 1 || len(resp.SiblingProjects) != 1 {
		t.Fatalf("unexpected siblings: %+v", resp.SiblingProjects)
	}
}

func TestListSiblingProjectsMissingParent(t *testing.T) {
	_, err := listSiblingProjects("/does/not/exist", "/does/not/exist/child")
	if err == nil {
		t.Fatal("expected error for missing parent directory")
	}
}
