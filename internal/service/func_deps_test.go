package service

import (
	"slices"
	"testing"
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
	writeFile(t, root+"/go.mod", "module demo\n\nrequire github.com/foo/bar v1.0.0\n")
	writeFile(t, root+"/package.json", `{"dependencies":{"left-pad":"1.0.0"}}`)

	chdir(t, root)

	jsonText, err := Dependencies()
	if err != nil {
		t.Fatalf("Dependencies() error: %v", err)
	}

	var resp dependenciesResponse
	parseJSONResponse(t, jsonText, &resp)

	if resp.ToolName != "dependencies" || resp.EcosystemCount != 2 {
		t.Fatalf("unexpected response: ecosystemCount=%d ecosystems=%+v", resp.EcosystemCount, resp.Ecosystems)
	}
	if !slices.Contains(resp.DetectedEcosystemNames, "go") || !slices.Contains(resp.DetectedEcosystemNames, "node") {
		t.Fatalf("detected ecosystems = %v", resp.DetectedEcosystemNames)
	}
}
