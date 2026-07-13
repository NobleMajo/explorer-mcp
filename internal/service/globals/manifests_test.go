package globals

import (
	"path/filepath"
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

	byName := make(map[string]GoDependency, len(deps))
	for _, dep := range deps {
		byName[dep.PackageName] = dep
	}

	foo, ok := byName["github.com/foo/bar"]
	if !ok || foo.Version != "v1.2.3" || foo.IsIndirect {
		t.Fatalf("unexpected foo dep: %+v", foo)
	}

	indirect, ok := byName["github.com/indirect/dep"]
	if !ok || !indirect.IsIndirect {
		t.Fatalf("expected indirect dep: %+v", indirect)
	}

	single, ok := byName["github.com/single/dep"]
	if !ok || single.Version != "v9.9.9" {
		t.Fatalf("unexpected single dep: %+v", single)
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

func TestLoadGoModManifestMissingFile(t *testing.T) {
	t.Parallel()

	_, err := LoadGoModManifest(t.TempDir(), "/does/not/exist/go.mod")
	if err == nil {
		t.Fatal("expected error for missing go.mod")
	}
}

func TestLoadPackageJsonManifestMissingFile(t *testing.T) {
	t.Parallel()

	_, err := LoadPackageJsonManifest(t.TempDir(), "/does/not/exist/package.json")
	if err == nil {
		t.Fatal("expected error for missing package.json")
	}
}

func TestLoadPackageJsonManifestInvalidJSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "package.json")
	testutil.WriteFile(t, path, `{invalid`)

	_, err := LoadPackageJsonManifest(root, path)
	if err == nil {
		t.Fatal("expected error for invalid package.json")
	}
}

func TestLoadGoModManifest(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "go.mod")
	testutil.WriteFile(t, path, "module test\n\nrequire github.com/foo/bar v1.0.0\n")

	got, err := LoadGoModManifest(root, path)
	if err != nil {
		t.Fatalf("LoadGoModManifest() error: %v", err)
	}
	if got.EcosystemName != "go" || !got.IsParsed || got.DependencyCount != 1 {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestLoadPackageJsonManifest(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "package.json")
	testutil.WriteFile(t, path, `{"dependencies":{"alpha":"1.0.0"},"devDependencies":{"eslint":"9.0.0"}}`)

	got, err := LoadPackageJsonManifest(root, path)
	if err != nil {
		t.Fatalf("LoadPackageJsonManifest() error: %v", err)
	}
	if got.EcosystemName != "node" || !got.IsParsed || len(got.DependencyGroups) != 2 {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestLoadRequirementsManifest(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "requirements.txt")
	testutil.WriteFile(t, path, "# comment\nrequests==2.28.0\nflask>=3.0.0\n\n")

	got, err := LoadRequirementsManifest(root, path)
	if err != nil {
		t.Fatalf("LoadRequirementsManifest() error: %v", err)
	}
	if !got.IsParsed || len(got.DependencyGroups) != 1 {
		t.Fatalf("unexpected groups: %+v", got.DependencyGroups)
	}
	names := got.DependencyGroups[0].PackageNames
	if len(names) != 2 || names[0] != "flask" || names[1] != "requests" {
		t.Fatalf("unexpected package names: %#v", names)
	}
}

func TestLoadRequirementsManifestMissingFile(t *testing.T) {
	t.Parallel()

	_, err := LoadRequirementsManifest(t.TempDir(), "/does/not/exist/requirements.txt")
	if err == nil {
		t.Fatal("expected error for missing requirements.txt")
	}
}

func TestLoadRequirementsManifestInlineComment(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "requirements.txt")
	testutil.WriteFile(t, path, "requests==2.28.0 # pinned\n")

	got, err := LoadRequirementsManifest(root, path)
	if err != nil {
		t.Fatalf("LoadRequirementsManifest() error: %v", err)
	}
	names := got.DependencyGroups[0].PackageNames
	if len(names) != 1 || names[0] != "requests" {
		t.Fatalf("unexpected package names: %#v", names)
	}
}

func TestLoadRequirementsManifestPlainPackageName(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "requirements.txt")
	testutil.WriteFile(t, path, "requests\n")

	got, err := LoadRequirementsManifest(root, path)
	if err != nil {
		t.Fatalf("LoadRequirementsManifest() error: %v", err)
	}
	names := got.DependencyGroups[0].PackageNames
	if len(names) != 1 || names[0] != "requests" {
		t.Fatalf("unexpected package names: %#v", names)
	}
}

func TestLoadRequirementsManifestEmptyFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "requirements.txt")
	testutil.WriteFile(t, path, "")

	got, err := LoadRequirementsManifest(root, path)
	if err != nil {
		t.Fatalf("LoadRequirementsManifest() error: %v", err)
	}
	if !got.IsParsed || len(got.DependencyGroups) != 1 || len(got.DependencyGroups[0].PackageNames) != 0 {
		t.Fatalf("unexpected empty requirements result: %+v", got)
	}
}

func TestLoadCargoManifestDetectOnly(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "Cargo.toml")
	testutil.WriteFile(t, path, "[package]\nname = \"demo\"\n")

	got, err := LoadCargoManifest(root, path)
	if err != nil {
		t.Fatalf("LoadCargoManifest() error: %v", err)
	}
	if got.IsParsed || got.EcosystemName != "rust" || got.ParseSkipReason == "" {
		t.Fatalf("unexpected detect-only result: %+v", got)
	}
}

func TestLoadPyprojectManifestDetectOnly(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "pyproject.toml")
	testutil.WriteFile(t, path, "[project]\nname = \"demo\"\n")

	got, err := LoadPyprojectManifest(root, path)
	if err != nil {
		t.Fatalf("LoadPyprojectManifest() error: %v", err)
	}
	if got.IsParsed || got.EcosystemName != "python" {
		t.Fatalf("unexpected detect-only result: %+v", got)
	}
}

func TestEcosystemNameForManifest(t *testing.T) {
	t.Parallel()

	cases := []struct {
		path string
		want string
	}{
		{path: "go.mod", want: "go"},
		{path: "package.json", want: "node"},
		{path: "requirements.txt", want: "python"},
		{path: "Cargo.toml", want: "rust"},
		{path: "pyproject.toml", want: "python"},
		{path: "unknown.lock", want: "unknown"},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()
			if got := ecosystemNameForManifest(tc.path); got != tc.want {
				t.Fatalf("ecosystemNameForManifest(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}
