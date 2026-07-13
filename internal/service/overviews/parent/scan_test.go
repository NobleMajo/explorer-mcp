package parent

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func scanSettings(depth int) ScanSettings {
	return ScanSettings{Depth: depth}
}

func TestListParentProjectsDepthOne(t *testing.T) {
	parent := t.TempDir()
	current := filepath.Join(parent, "app")
	sibling := filepath.Join(parent, "other")
	nested := filepath.Join(sibling, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, filepath.Join(parent, "notes.txt"), "ignore\n")

	siblings, err := listParentProjects(current, scanSettings(1))
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

	siblings, err := listParentProjects(current, scanSettings(2))
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
	if err := os.MkdirAll(filepath.Join(gitSibling, "internal", "pkg"), 0o755); err != nil {
		t.Fatal(err)
	}

	siblings, err := listParentProjects(current, scanSettings(3))
	if err != nil {
		t.Fatalf("listParentProjects() error: %v", err)
	}
	if !slices.Contains(siblings, "../beta @git") {
		t.Fatalf("missing ../beta @git in siblings %v", siblings)
	}
	for _, entry := range siblings {
		if strings.HasPrefix(siblingRelativePath(entry), "../beta/") {
			t.Fatalf("expected no nested paths under flagged repo, got %q", entry)
		}
	}
}

func TestListParentProjectsSkipsNestedDirsWhenFlagged(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	current := filepath.Join(parent, "app")
	goSibling := filepath.Join(parent, "wgg")
	if err := os.MkdirAll(current, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(goSibling, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, filepath.Join(goSibling, "go.mod"), "module demo\n")
	for _, nested := range []string{"internal", "lib", ".github"} {
		if err := os.MkdirAll(filepath.Join(goSibling, nested), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	siblings, err := listParentProjects(current, scanSettings(3))
	if err != nil {
		t.Fatalf("listParentProjects() error: %v", err)
	}
	want := "../wgg @git @go"
	if !slices.Contains(siblings, want) {
		t.Fatalf("missing %q in siblings %v", want, siblings)
	}
	for _, entry := range siblings {
		if strings.HasPrefix(siblingRelativePath(entry), "../wgg/") {
			t.Fatalf("expected no nested paths under flagged project, got %q", entry)
		}
	}
}

func TestListParentProjectsSkipsNestedDirsForMakefileProject(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	current := filepath.Join(parent, "app")
	makeSibling := filepath.Join(parent, "tools")
	if err := os.MkdirAll(current, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, filepath.Join(makeSibling, "Makefile"), "build:\n")
	if err := os.MkdirAll(filepath.Join(makeSibling, "cmd"), 0o755); err != nil {
		t.Fatal(err)
	}

	siblings, err := listParentProjects(current, scanSettings(3))
	if err != nil {
		t.Fatalf("listParentProjects() error: %v", err)
	}
	want := "../tools @makefile"
	if !slices.Contains(siblings, want) {
		t.Fatalf("missing %q in siblings %v", want, siblings)
	}
	for _, entry := range siblings {
		if strings.HasPrefix(siblingRelativePath(entry), "../tools/") {
			t.Fatalf("expected no nested paths under makefile project, got %q", entry)
		}
	}
}

func TestListParentProjectsSkipsNestedDirsForDockerComposeProject(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	current := filepath.Join(parent, "app")
	composeSibling := filepath.Join(parent, "stack")
	if err := os.MkdirAll(current, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, filepath.Join(composeSibling, "docker-compose.yml"), "services: {}\n")
	if err := os.MkdirAll(filepath.Join(composeSibling, "services"), 0o755); err != nil {
		t.Fatal(err)
	}

	siblings, err := listParentProjects(current, scanSettings(3))
	if err != nil {
		t.Fatalf("listParentProjects() error: %v", err)
	}
	want := "../stack @docker-compose"
	if !slices.Contains(siblings, want) {
		t.Fatalf("missing %q in siblings %v", want, siblings)
	}
	for _, entry := range siblings {
		if strings.HasPrefix(siblingRelativePath(entry), "../stack/") {
			t.Fatalf("expected no nested paths under compose project, got %q", entry)
		}
	}
}

func TestListParentProjectsSkipsNestedDirsForGNUmakefileProject(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	current := filepath.Join(parent, "app")
	makeSibling := filepath.Join(parent, "legacy")
	if err := os.MkdirAll(current, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, filepath.Join(makeSibling, "GNUmakefile"), "build:\n")
	if err := os.MkdirAll(filepath.Join(makeSibling, "src"), 0o755); err != nil {
		t.Fatal(err)
	}

	siblings, err := listParentProjects(current, scanSettings(3))
	if err != nil {
		t.Fatalf("listParentProjects() error: %v", err)
	}
	want := "../legacy @makefile"
	if !slices.Contains(siblings, want) {
		t.Fatalf("missing %q in siblings %v", want, siblings)
	}
	for _, entry := range siblings {
		if strings.HasPrefix(siblingRelativePath(entry), "../legacy/") {
			t.Fatalf("expected no nested paths under makefile project, got %q", entry)
		}
	}
}

func TestListParentProjectsSkipsCurrentProjectSubtree(t *testing.T) {
	root := t.TempDir()
	current := filepath.Join(root, "app")
	inside := filepath.Join(current, "pkg")
	if err := os.MkdirAll(inside, 0o755); err != nil {
		t.Fatal(err)
	}

	siblings, err := listParentProjects(current, scanSettings(2))
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
	_, err := listParentProjects("/does/not/exist", scanSettings(1))
	if err == nil {
		t.Fatal("expected error for missing start directory")
	}
}

func TestListParentProjectsSkipsDotDirsByDefault(t *testing.T) {
	root := t.TempDir()
	current := filepath.Join(root, "app")
	dotSibling := filepath.Join(root, ".config")
	visibleSibling := filepath.Join(root, "visible")
	for _, dir := range []string{current, dotSibling, visibleSibling} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	siblings, err := listParentProjects(current, ScanSettings{Depth: 1})
	if err != nil {
		t.Fatalf("listParentProjects() error: %v", err)
	}
	if !slices.ContainsFunc(siblings, func(entry string) bool {
		return siblingRelativePath(entry) == "../visible"
	}) {
		t.Fatalf("missing ../visible in siblings %v", siblings)
	}
	for _, entry := range siblings {
		if strings.Contains(siblingRelativePath(entry), ".config") {
			t.Fatalf("expected dot dir skipped, got %q", entry)
		}
	}
}

func TestListParentProjectsIncludesDotDirsWithFlag(t *testing.T) {
	root := t.TempDir()
	current := filepath.Join(root, "app")
	dotSibling := filepath.Join(root, ".config")
	if err := os.MkdirAll(current, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dotSibling, 0o755); err != nil {
		t.Fatal(err)
	}

	siblings, err := listParentProjects(current, ScanSettings{Depth: 1, ScanDotDirs: true})
	if err != nil {
		t.Fatalf("listParentProjects() error: %v", err)
	}
	if !slices.ContainsFunc(siblings, func(entry string) bool {
		return siblingRelativePath(entry) == "../.config"
	}) {
		t.Fatalf("missing ../.config in siblings %v", siblings)
	}
}

func TestListParentProjectsStopsAtHomeByDefault(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home", "user")
	current := filepath.Join(home, "ws", "app")
	homeSibling := filepath.Join(home, "other")
	wsSibling := filepath.Join(home, "ws", "sib")
	for _, dir := range []string{current, homeSibling, wsSibling} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	siblings, err := listParentProjects(current, ScanSettings{
		Depth:   3,
		HomeDir: home,
	})
	if err != nil {
		t.Fatalf("listParentProjects() error: %v", err)
	}
	if !slices.ContainsFunc(siblings, func(entry string) bool {
		return siblingRelativePath(entry) == "../sib"
	}) {
		t.Fatalf("missing ../sib in siblings %v", siblings)
	}
	for _, entry := range siblings {
		rel := siblingRelativePath(entry)
		if rel == "../../other" || strings.HasPrefix(rel, "../../other/") {
			t.Fatalf("expected home scan to stop before listing %q", rel)
		}
	}
}

func TestListParentProjectsScansHomeWithFlag(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home", "user")
	current := filepath.Join(home, "ws", "app")
	homeSibling := filepath.Join(home, "other")
	wsSibling := filepath.Join(home, "ws", "sib")
	for _, dir := range []string{current, homeSibling, wsSibling} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	siblings, err := listParentProjects(current, ScanSettings{
		Depth:       3,
		ScanHomeDir: true,
		HomeDir:     home,
	})
	if err != nil {
		t.Fatalf("listParentProjects() error: %v", err)
	}
	for _, want := range []string{"../sib", "../../other"} {
		if !slices.ContainsFunc(siblings, func(entry string) bool {
			return siblingRelativePath(entry) == want
		}) {
			t.Fatalf("missing %q in siblings %v", want, siblings)
		}
	}
}

func TestShouldStopParentScanAtFilesystemRoot(t *testing.T) {
	ctx := scanContext{settings: ScanSettings{Depth: 5}}
	if !shouldStopParentScan("/", "/", ctx) {
		t.Fatal("expected stop at filesystem root")
	}
	if !shouldStopParentScan("/", "/tmp", ctx) {
		t.Fatal("expected stop before scanning from filesystem root")
	}
}

func TestShouldStopParentScanAtHome(t *testing.T) {
	home := filepath.Clean("/home/user")
	ctx := scanContext{
		settings: ScanSettings{Depth: 5, ScanHomeDir: false, HomeDir: home},
		homeDir:  home,
	}
	if !shouldStopParentScan(home, filepath.Join(home, "ws", "app"), ctx) {
		t.Fatal("expected stop at home directory boundary")
	}
}
