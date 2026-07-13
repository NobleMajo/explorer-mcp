package deps

import (
	"slices"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestContainsEcosystemName(t *testing.T) {
	t.Parallel()

	names := []string{"go", "node"}
	if !containsEcosystemName(names, "go") {
		t.Fatal("expected go to be found")
	}
	if containsEcosystemName(names, "python") {
		t.Fatal("expected python to be missing")
	}
}

func TestDependenciesFindsManifests(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/go.mod", "module demo\n\nrequire github.com/foo/bar v1.0.0\n")
	testutil.WriteFile(t, root+"/package.json", `{"dependencies":{"left-pad":"1.0.0"}}`)

	testutil.Chdir(t, root)

	result, err := DepsOverview()(false)
	if err != nil {
		t.Fatalf("DepsOverview() error: %v", err)
	}

	resp, ok := result.(dependenciesResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	if resp.ToolName != "dependencies" || resp.EcosystemCount != 2 {
		t.Fatalf("unexpected response: ecosystemCount=%d ecosystems=%+v", resp.EcosystemCount, resp.Ecosystems)
	}
	if !slices.Contains(resp.DetectedEcosystemNames, "go") || !slices.Contains(resp.DetectedEcosystemNames, "node") {
		t.Fatalf("detected ecosystems = %v", resp.DetectedEcosystemNames)
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

	resp, ok := result.(dependenciesResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	if resp.EcosystemCount != 2 {
		t.Fatalf("ecosystemCount = %d, want 2", resp.EcosystemCount)
	}
	for _, eco := range resp.Ecosystems {
		if eco.IsParsed {
			t.Fatalf("expected detect-only for %s, got %+v", eco.EcosystemName, eco)
		}
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

	resp, ok := result.(dependenciesResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if resp.EcosystemCount != 1 || resp.Ecosystems[0].EcosystemName != "python" {
		t.Fatalf("unexpected python deps: %+v", resp)
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

	resp, ok := result.(dependenciesResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	if resp.EcosystemCount != len(globals.ManifestLoaders) {
		t.Fatalf("ecosystemCount = %d, want %d", resp.EcosystemCount, len(globals.ManifestLoaders))
	}

	manifestPaths := make([]string, 0, len(resp.Ecosystems))
	for _, eco := range resp.Ecosystems {
		manifestPaths = append(manifestPaths, eco.ManifestFilePath)
	}
	for fileName := range globals.ManifestLoaders {
		if !slices.Contains(manifestPaths, fileName) {
			t.Fatalf("missing manifest %q in ecosystems: %+v", fileName, resp.Ecosystems)
		}
	}

	wantNames := []string{"go", "node", "python", "rust"}
	if len(resp.DetectedEcosystemNames) != len(wantNames) {
		t.Fatalf("detected ecosystems = %v, want %v", resp.DetectedEcosystemNames, wantNames)
	}
	for _, name := range wantNames {
		if !slices.Contains(resp.DetectedEcosystemNames, name) {
			t.Fatalf("missing ecosystem name %q in %v", name, resp.DetectedEcosystemNames)
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

	resp, ok := result.(dependenciesResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if resp.EcosystemCount != 0 || len(resp.Ecosystems) != 0 {
		t.Fatalf("expected no ecosystems, got %+v", resp)
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

	resp, ok := result.(dependenciesResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if resp.EcosystemCount != 1 || resp.Ecosystems[0].EcosystemName != "go" {
		t.Fatalf("expected parsed go ecosystem, got %+v", resp)
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
