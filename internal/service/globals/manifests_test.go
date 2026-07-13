package globals

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestParseGoModRequire(t *testing.T) {
	t.Parallel()

	content := `
module example.com/app

require (
	github.com/foo/bar v1.2.3
	github.com/indirect/dep v0.1.0 // indirect
)

require github.com/single/dep v9.9.9
`

	deps := parseGoModRequire(content)
	if len(deps) != 3 {
		t.Fatalf("len(deps) = %d, want 3", len(deps))
	}

	want := []string{
		"github.com/foo/bar@v1.2.3 direct",
		"github.com/indirect/dep@v0.1.0 indirect",
		"github.com/single/dep@v9.9.9 direct",
	}
	for _, entry := range want {
		if !slices.Contains(deps, entry) {
			t.Fatalf("missing %q in %v", entry, deps)
		}
	}
}

func TestSortedManifestKeys(t *testing.T) {
	t.Parallel()

	keys := sortedManifestKeys(map[string]string{
		"zeta":   "1",
		"alpha":  "2",
		"middle": "3",
	})

	want := []string{"alpha", "middle", "zeta"}
	if len(keys) != len(want) {
		t.Fatalf("len(keys) = %d, want %d", len(keys), len(want))
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Fatalf("keys[%d] = %q, want %q", i, keys[i], want[i])
		}
	}
}

func TestParseGoModRequireLineInvalid(t *testing.T) {
	t.Parallel()

	_, ok := parseGoModRequireLine("incomplete")
	if ok {
		t.Fatal("expected invalid require line to fail")
	}
}

func TestParseGoModTool(t *testing.T) {
	t.Parallel()

	content := `module demo

require golang.org/x/tools v0.30.0

tool (
	golang.org/x/tools/cmd/goimports
)

tool github.com/golangci/golangci-lint/cmd/golangci-lint
`
	versions := parseGoModRequireVersions(content)
	got := parseGoModTool(content, versions)
	want := []string{
		"golang.org/x/tools/cmd/goimports@v0.30.0 tool",
		"github.com/golangci/golangci-lint/cmd/golangci-lint tool",
	}
	if len(got) != len(want) {
		t.Fatalf("parseGoModTool() = %v, want %v", got, want)
	}
	for _, entry := range want {
		if !slices.Contains(got, entry) {
			t.Fatalf("missing %q in %v", entry, got)
		}
	}
}

func TestParseGoModDependenciesHidesToolDepsWhenDisabled(t *testing.T) {
	t.Parallel()

	content := `module demo

require golang.org/x/tools v0.30.0

tool golang.org/x/tools/cmd/goimports
`
	got := parseGoModDependencies(content, ManifestDepsSettings{ShowGoToolDeps: false})
	for _, entry := range got {
		if strings.HasSuffix(entry, " tool") {
			t.Fatalf("unexpected tool dependency %q", entry)
		}
	}
	if len(got) != 1 || got[0] != "golang.org/x/tools@v0.30.0 direct" {
		t.Fatalf("unexpected result: %v", got)
	}
}

func TestLoadGoModManifestIncludesToolDepsByDefault(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "go.mod")
	testutil.WriteFile(t, path, `module demo

require golang.org/x/tools v0.30.0

tool golang.org/x/tools/cmd/goimports
`)

	tag, got, loaded, err := LoadGoModManifest(root, path)
	if err != nil {
		t.Fatalf("LoadGoModManifest() error: %v", err)
	}
	if !loaded || tag != "@go" {
		t.Fatalf("loaded=%v tag=%q, want loaded true tag @go", loaded, tag)
	}
	if !slices.Contains(got, "golang.org/x/tools/cmd/goimports@v0.30.0 tool") {
		t.Fatalf("missing tool dependency, got %v", got)
	}
}

func TestLoadGoModManifestMissingFile(t *testing.T) {
	t.Parallel()

	_, _, loaded, err := LoadGoModManifest(t.TempDir(), "/does/not/exist/go.mod")
	if err != nil {
		t.Fatalf("LoadGoModManifest() error: %v", err)
	}
	if loaded {
		t.Fatal("expected loaded false for missing go.mod")
	}
}

func TestLoadPackageJsonManifestMissingFile(t *testing.T) {
	t.Parallel()

	_, _, loaded, err := LoadPackageJsonManifest(t.TempDir(), "/does/not/exist/package.json")
	if err != nil {
		t.Fatalf("LoadPackageJsonManifest() error: %v", err)
	}
	if loaded {
		t.Fatal("expected loaded false for missing package.json")
	}
}

func TestLoadPackageJsonManifestInvalidJSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "package.json")
	testutil.WriteFile(t, path, `{invalid`)

	_, _, loaded, err := LoadPackageJsonManifest(root, path)
	if err == nil {
		t.Fatal("expected error for invalid package.json")
	}
	if loaded {
		t.Fatal("expected loaded false for invalid package.json")
	}
}

func TestLoadGoModManifest(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "go.mod")
	testutil.WriteFile(t, path, "module test\n\nrequire github.com/foo/bar v1.0.0\n")

	tag, got, loaded, err := LoadGoModManifest(root, path)
	if err != nil {
		t.Fatalf("LoadGoModManifest() error: %v", err)
	}
	if !loaded {
		t.Fatal("expected loaded true")
	}
	if tag != "@go" {
		t.Fatalf("tag = %q, want @go", tag)
	}
	if len(got) != 1 || got[0] != "github.com/foo/bar@v1.0.0 direct" {
		t.Fatalf("unexpected result: %v", got)
	}
}

func TestLoadPackageJsonManifest(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "package.json")
	testutil.WriteFile(t, path, `{"dependencies":{"alpha":"1.0.0"},"devDependencies":{"eslint":"9.0.0"}}`)

	tag, got, loaded, err := LoadPackageJsonManifest(root, path)
	if err != nil {
		t.Fatalf("LoadPackageJsonManifest() error: %v", err)
	}
	if !loaded {
		t.Fatal("expected loaded true")
	}
	if tag != "@npm" {
		t.Fatalf("tag = %q, want @npm", tag)
	}
	want := []string{"alpha@1.0.0 direct", "eslint@9.0.0 dev"}
	if len(got) != len(want) {
		t.Fatalf("unexpected result: %v", got)
	}
	for _, entry := range want {
		if !slices.Contains(got, entry) {
			t.Fatalf("missing %q in %v", entry, got)
		}
	}
}

func TestLoadRequirementsManifest(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "requirements.txt")
	testutil.WriteFile(t, path, "# comment\nrequests==2.28.0\nflask>=3.0.0\n\n")

	tag, got, loaded, err := LoadRequirementsManifest(root, path)
	if err != nil {
		t.Fatalf("LoadRequirementsManifest() error: %v", err)
	}
	if !loaded {
		t.Fatal("expected loaded true")
	}
	if tag != "@pip" {
		t.Fatalf("tag = %q, want @pip", tag)
	}
	want := []string{"flask@>=3.0.0 direct", "requests@==2.28.0 direct"}
	if len(got) != len(want) {
		t.Fatalf("unexpected result: %v", got)
	}
	for _, entry := range want {
		if !slices.Contains(got, entry) {
			t.Fatalf("missing %q in %v", entry, got)
		}
	}
}

func TestLoadRequirementsManifestMissingFile(t *testing.T) {
	t.Parallel()

	_, _, loaded, err := LoadRequirementsManifest(t.TempDir(), "/does/not/exist/requirements.txt")
	if err != nil {
		t.Fatalf("LoadRequirementsManifest() error: %v", err)
	}
	if loaded {
		t.Fatal("expected loaded false for missing requirements.txt")
	}
}

func TestLoadRequirementsManifestInlineComment(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "requirements.txt")
	testutil.WriteFile(t, path, "requests==2.28.0 # pinned\n")

	tag, got, loaded, err := LoadRequirementsManifest(root, path)
	if err != nil {
		t.Fatalf("LoadRequirementsManifest() error: %v", err)
	}
	if !loaded {
		t.Fatal("expected loaded true")
	}
	if tag != "@pip" {
		t.Fatalf("tag = %q, want @pip", tag)
	}
	if len(got) != 1 || got[0] != "requests@==2.28.0 direct" {
		t.Fatalf("unexpected result: %v", got)
	}
}

func TestLoadRequirementsManifestPlainPackageName(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "requirements.txt")
	testutil.WriteFile(t, path, "requests\n")

	tag, got, loaded, err := LoadRequirementsManifest(root, path)
	if err != nil {
		t.Fatalf("LoadRequirementsManifest() error: %v", err)
	}
	if !loaded {
		t.Fatal("expected loaded true")
	}
	if tag != "@pip" {
		t.Fatalf("tag = %q, want @pip", tag)
	}
	if len(got) != 1 || got[0] != "requests direct" {
		t.Fatalf("unexpected result: %v", got)
	}
}

func TestLoadRequirementsManifestEmptyFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "requirements.txt")
	testutil.WriteFile(t, path, "")

	tag, got, loaded, err := LoadRequirementsManifest(root, path)
	if err != nil {
		t.Fatalf("LoadRequirementsManifest() error: %v", err)
	}
	if !loaded {
		t.Fatal("expected loaded true")
	}
	if tag != "@pip" {
		t.Fatalf("tag = %q, want @pip", tag)
	}
	if len(got) != 0 {
		t.Fatalf("unexpected empty requirements result: %v", got)
	}
}

func TestLoadCargoManifestDetectOnly(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "Cargo.toml")
	testutil.WriteFile(t, path, "[package]\nname = \"demo\"\n")

	tag, got, loaded, err := LoadCargoManifest(root, path)
	if err != nil {
		t.Fatalf("LoadCargoManifest() error: %v", err)
	}
	if !loaded {
		t.Fatal("expected loaded true")
	}
	if tag != "@cargo" {
		t.Fatalf("tag = %q, want @cargo", tag)
	}
	if len(got) != 0 {
		t.Fatalf("expected no parsed cargo dependencies, got %v", got)
	}
}

func TestLoadPyprojectManifestDetectOnly(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "pyproject.toml")
	testutil.WriteFile(t, path, "[project]\nname = \"demo\"\n")

	tag, got, loaded, err := LoadPyprojectManifest(root, path)
	if err != nil {
		t.Fatalf("LoadPyprojectManifest() error: %v", err)
	}
	if !loaded {
		t.Fatal("expected loaded true")
	}
	if tag != "@python" {
		t.Fatalf("tag = %q, want @python", tag)
	}
	if len(got) != 0 {
		t.Fatalf("expected no parsed pyproject dependencies, got %v", got)
	}
}
