package parent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestFormatSiblingProject(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	plain := filepath.Join(root, "plain")
	gitOnly := filepath.Join(root, "git-only")
	goGit := filepath.Join(root, "go-git")
	for _, dir := range []string{plain, gitOnly, goGit} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(gitOnly, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(goGit, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, filepath.Join(goGit, "go.mod"), "module demo\n")

	tests := []struct {
		name    string
		absPath string
		relPath string
		want    string
	}{
		{name: "plain", absPath: plain, relPath: "../plain", want: "../plain"},
		{name: "git only", absPath: gitOnly, relPath: "../git-only", want: "../git-only @git"},
		{name: "git and go", absPath: goGit, relPath: "../go-git", want: "../go-git @git @go"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := formatSiblingProject(tc.absPath, tc.relPath); got != tc.want {
				t.Fatalf("formatSiblingProject() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatSiblingProjectManifestTags(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	npmDir := filepath.Join(root, "npm-app")
	if err := os.MkdirAll(npmDir, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, filepath.Join(npmDir, "package.json"), "{}\n")

	got := formatSiblingProject(npmDir, "../npm-app")
	want := "../npm-app @npm"
	if got != want {
		t.Fatalf("formatSiblingProject() = %q, want %q", got, want)
	}
}
