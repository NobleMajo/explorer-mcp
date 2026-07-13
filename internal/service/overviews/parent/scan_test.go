package parent

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestListParentProjectsDepthOne(t *testing.T) {
	parent := t.TempDir()
	current := filepath.Join(parent, "app")
	sibling := filepath.Join(parent, "other")
	nested := filepath.Join(sibling, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, filepath.Join(parent, "notes.txt"), "ignore\n")

	siblings, err := listParentProjects(current, 1)
	if err != nil {
		t.Fatalf("listParentProjects() error: %v", err)
	}
	if len(siblings) != 1 || siblingRelativePath(siblings[0]) != "../other" {
		t.Fatalf("depth 1 = %v, want [../other]", siblings)
	}
}

func TestListParentProjectsDepthTwo(t *testing.T) {
	root := t.TempDir()
	grand := filepath.Join(root, "grand")
	parent := filepath.Join(grand, "parent")
	current := filepath.Join(parent, "app")
	sibling := filepath.Join(parent, "sib")
	nested := filepath.Join(sibling, "nested")
	cousin := filepath.Join(grand, "cousin")
	for _, dir := range []string{current, nested, cousin} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	siblings, err := listParentProjects(current, 2)
	if err != nil {
		t.Fatalf("listParentProjects() error: %v", err)
	}
	for _, want := range []string{"../sib", "../sib/nested", "../../cousin"} {
		if !slices.ContainsFunc(siblings, func(entry string) bool {
			return siblingRelativePath(entry) == want
		}) {
			t.Fatalf("missing %q in siblings %v", want, siblings)
		}
	}
}

func TestListParentProjectsDepthThreeIncludesGitRepo(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	current := filepath.Join(parent, "app")
	gitSibling := filepath.Join(parent, "beta")
	if err := os.MkdirAll(current, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(gitSibling, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	siblings, err := listParentProjects(current, 3)
	if err != nil {
		t.Fatalf("listParentProjects() error: %v", err)
	}
	if !slices.Contains(siblings, "../beta @git") {
		t.Fatalf("missing ../beta @git in siblings %v", siblings)
	}
}

func TestListParentProjectsSkipsCurrentProjectSubtree(t *testing.T) {
	root := t.TempDir()
	current := filepath.Join(root, "app")
	inside := filepath.Join(current, "pkg")
	if err := os.MkdirAll(inside, 0o755); err != nil {
		t.Fatal(err)
	}

	siblings, err := listParentProjects(current, 2)
	if err != nil {
		t.Fatalf("listParentProjects() error: %v", err)
	}
	for _, entry := range siblings {
		path := siblingRelativePath(entry)
		if path == "pkg" || path == "./pkg" {
			t.Fatalf("should not include cwd subtree %q", path)
		}
	}
}

func TestListParentProjectsMissingStartDir(t *testing.T) {
	_, err := listParentProjects("/does/not/exist", 1)
	if err == nil {
		t.Fatal("expected error for missing start directory")
	}
}
