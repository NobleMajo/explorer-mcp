package deps

import (
	"slices"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestDependenciesFindsManifests(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/go.mod", "module demo\n\nrequire github.com/foo/bar v1.0.0\n")
	testutil.WriteFile(t, root+"/package.json", `{"dependencies":{"left-pad":"1.0.0"}}`)

	testutil.Chdir(t, root)

	result, err := DepsOverview()(false)
	if err != nil {
		t.Fatalf("DepsOverview() error: %v", err)
	}

	resp, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	if !slices.Contains(resp, "github.com/foo/bar@v1.0.0 direct") {
		t.Fatalf("missing go dependency, got %v", resp)
	}
	if !slices.Contains(resp, "left-pad@1.0.0 production") {
		t.Fatalf("missing node dependency, got %v", resp)
	}
}

func TestDependenciesDetectOnlyManifests(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/Cargo.toml", "[package]\nname = \"demo\"\n")
	testutil.WriteFile(t, root+"/pyproject.toml", "[project]\nname = \"demo\"\n")

	testutil.Chdir(t, root)

	result, err := DepsOverview()(false)
	if err != nil {
		t.Fatalf("DepsOverview() error: %v", err)
	}

	resp, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if len(resp) != 0 {
		t.Fatalf("expected no parsed dependencies, got %v", resp)
	}
}

func TestDependenciesRequirementsTxt(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/requirements.txt", "flask>=3.0.0\n")
	testutil.Chdir(t, root)

	result, err := DepsOverview()(false)
	if err != nil {
		t.Fatalf("DepsOverview() error: %v", err)
	}

	resp, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if len(resp) != 1 || resp[0] != "flask@>=3.0.0" {
		t.Fatalf("unexpected python deps: %v", resp)
	}
}

func TestDependenciesAllManifestLoaders(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/go.mod", "module demo\n\nrequire github.com/foo/bar v1.0.0\n")
	testutil.WriteFile(t, root+"/package.json", `{"dependencies":{"left-pad":"1.0.0"}}`)
	testutil.WriteFile(t, root+"/requirements.txt", "flask>=3.0.0\n")
	testutil.WriteFile(t, root+"/Cargo.toml", "[package]\nname = \"demo\"\n")
	testutil.WriteFile(t, root+"/pyproject.toml", "[project]\nname = \"demo\"\n")
	testutil.Chdir(t, root)

	result, err := DepsOverview()(false)
	if err != nil {
		t.Fatalf("DepsOverview() error: %v", err)
	}

	resp, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	for _, want := range []string{
		"github.com/foo/bar@v1.0.0 direct",
		"left-pad@1.0.0 production",
		"flask@>=3.0.0",
	} {
		if !slices.Contains(resp, want) {
			t.Fatalf("missing dependency %q, got %v", want, resp)
		}
	}
}

func TestDependenciesSkipsMissingManifests(t *testing.T) {
	root := t.TempDir()
	testutil.Chdir(t, root)

	result, err := DepsOverview()(false)
	if err != nil {
		t.Fatalf("DepsOverview() error: %v", err)
	}

	resp, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if len(resp) != 0 {
		t.Fatalf("expected no dependencies, got %v", resp)
	}
}

func TestDependenciesGoModLoaderError(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/go.mod", "")
	testutil.Chdir(t, root)

	result, err := DepsOverview()(false)
	if err != nil {
		t.Fatalf("DepsOverview() error: %v", err)
	}

	resp, ok := result.([]string)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if len(resp) != 0 {
		t.Fatalf("expected empty go.mod to yield no dependencies, got %v", resp)
	}
}

func TestDependenciesLoaderError(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/package.json", `{invalid`)

	testutil.Chdir(t, root)

	_, err := DepsOverview()(false)
	if err == nil {
		t.Fatal("expected error for invalid package.json")
	}
}
