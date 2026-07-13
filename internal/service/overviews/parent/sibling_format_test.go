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
		name     string
		absPath  string
		relPath  string
		subfiles []string
		subdirs  []string
		want     string
	}{
		{name: "plain", absPath: plain, relPath: "../plain", want: "../plain"},
		{name: "git only", absPath: gitOnly, relPath: "../git-only", subdirs: []string{".git"}, want: "../git-only @git"},
		{name: "git and go", absPath: goGit, relPath: "../go-git", subfiles: []string{"go.mod"}, subdirs: []string{".git"}, want: "../go-git @git @go"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := formatSiblingProject(tc.absPath, tc.relPath, tc.subfiles, tc.subdirs); got != tc.want {
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

	got := formatSiblingProject(npmDir, "../npm-app", []string{"package.json"}, nil)
	want := "../npm-app @npm"
	if got != want {
		t.Fatalf("formatSiblingProject() = %q, want %q", got, want)
	}
}

func TestFormatSiblingProjectIdentifierTags(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	angularDir := filepath.Join(root, "web")
	if err := os.MkdirAll(angularDir, 0o755); err != nil {
		t.Fatal(err)
	}

	got := formatSiblingProject(angularDir, "../web", []string{"Makefile", "angular.json", "tsconfig.app.json"}, nil)
	want := "../web @makefile @tsconfig @angular"
	if got != want {
		t.Fatalf("formatSiblingProject() = %q, want %q", got, want)
	}
}

func TestHasSiblingProjectFlags(t *testing.T) {
	t.Parallel()

	if hasSiblingProjectFlags("/tmp/plain", nil, nil) {
		t.Fatal("expected plain dir to have no flags")
	}
	if !hasSiblingProjectFlags("/tmp/git", nil, []string{".git"}) {
		t.Fatal("expected git dir to have flags")
	}
	goDir := t.TempDir()
	testutil.WriteFile(t, filepath.Join(goDir, "go.mod"), "module demo\n")
	if !hasSiblingProjectFlags(goDir, []string{"go.mod"}, nil) {
		t.Fatal("expected go module dir to have flags")
	}
	angularDir := t.TempDir()
	testutil.WriteFile(t, filepath.Join(angularDir, "angular.json"), "{}\n")
	if !hasSiblingProjectFlags(angularDir, []string{"angular.json"}, nil) {
		t.Fatal("expected angular dir to have flags")
	}
}
