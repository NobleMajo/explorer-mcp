package container

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/service/globals"
	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

var globPatternFixtureFiles = map[string]string{
	"*.dockerfile":          "app.dockerfile",
	"*.Dockerfile":          "App.Dockerfile",
	"Dockerfile.*":          "Dockerfile.staging",
	"Containerfile.*":       "Containerfile.staging",
	"docker-compose.*.yml":  "docker-compose.override.yml",
	"docker-compose.*.yaml": "docker-compose.override.yaml",
	"compose.*.yml":         "compose.override.yml",
	"compose.*.yaml":        "compose.override.yaml",
}

func TestDetectContainerFilePathsCoversKnownGlobals(t *testing.T) {
	root := t.TempDir()

	expected := make([]string, 0, len(globals.KnownContainerFileNames)+len(globals.KnownContainerDirectoryNames)+1)

	for _, name := range globals.KnownContainerFileNames {
		if fixture, ok := globPatternFixtureFiles[name]; ok {
			testutil.WriteFile(t, filepath.Join(root, fixture), "fixture\n")
			expected = append(expected, fixture)
			continue
		}
		testutil.WriteFile(t, filepath.Join(root, name), "fixture\n")
		expected = append(expected, name)
	}

	for _, dirName := range globals.KnownContainerDirectoryNames {
		if err := os.Mkdir(filepath.Join(root, dirName), 0o755); err != nil {
			t.Fatal(err)
		}
		expected = append(expected, dirName+"/")
	}
	testutil.WriteFile(t, filepath.Join(root, ".devcontainer", "devcontainer.json"), "{}\n")
	expected = append(expected, ".devcontainer/devcontainer.json")

	paths, err := detectContainerFilePaths(root)
	if err != nil {
		t.Fatalf("detectContainerFilePaths() error: %v", err)
	}

	for _, want := range expected {
		if !slices.Contains(paths, want) {
			t.Fatalf("missing %q in paths: %v", want, paths)
		}
	}
	if len(paths) != len(expected) {
		t.Fatalf("path count = %d, want %d, got %v", len(paths), len(expected), paths)
	}
}

func TestContainerOverviewIntegration(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "Dockerfile"), "FROM alpine\n")
	testutil.Chdir(t, root)

	binDir := t.TempDir()
	for _, runtime := range []string{"docker", "podman"} {
		script := filepath.Join(binDir, runtime)
		testutil.WriteFile(t, script, "#!/bin/sh\nif [ \"$1\" = ps ]; then printf 'web\\timg:1\\tUp\\n'; fi\n")
		if err := os.Chmod(script, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	result, err := ContainerOverview()(false)
	if err != nil {
		t.Fatalf("ContainerOverview() error: %v", err)
	}

	resp, ok := result.(containerOverviewResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	if resp.ToolName != "container_overview" {
		t.Fatalf("toolName = %q", resp.ToolName)
	}
	if !resp.IsDockerAvailable || !resp.IsPodmanAvailable {
		t.Fatalf("expected both runtimes available, got docker=%v podman=%v", resp.IsDockerAvailable, resp.IsPodmanAvailable)
	}
	if resp.DetectedContainerFileCount != 1 {
		t.Fatalf("detectedContainerFileCount = %d, want 1", resp.DetectedContainerFileCount)
	}
	if resp.RunningContainerCount != 2 {
		t.Fatalf("runningContainerCount = %d, want 2", resp.RunningContainerCount)
	}
}

func TestListRunningContainers(t *testing.T) {
	t.Run("parses output", func(t *testing.T) {
		root := t.TempDir()
		fakeDocker := filepath.Join(root, "docker")
		testutil.WriteFile(t, fakeDocker, "#!/bin/sh\nif [ \"$1\" = ps ]; then\n  printf 'app\\tnginx:alpine\\tUp 2 hours\\n'\n  printf 'db\\tpostgres:15\\tUp 1 day\\n'\nfi\n")
		if err := os.Chmod(fakeDocker, 0o755); err != nil {
			t.Fatal(err)
		}

		t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))

		got := listRunningContainers("docker", root)
		if len(got) != 2 {
			t.Fatalf("len(containers) = %d, want 2, got %+v", len(got), got)
		}
		if got[0].ContainerName != "app" || got[0].ImageName != "nginx:alpine" || got[0].RuntimeName != "docker" {
			t.Fatalf("unexpected first container: %+v", got[0])
		}
		if got[1].ContainerName != "db" || got[1].StatusText != "Up 1 day" {
			t.Fatalf("unexpected second container: %+v", got[1])
		}
	})

	t.Run("exec error", func(t *testing.T) {
		got := listRunningContainers("definitely-not-a-runtime-cmd", t.TempDir())
		if len(got) != 0 {
			t.Fatalf("expected empty slice on exec error, got %#v", got)
		}
	})

	t.Run("empty line skipped", func(t *testing.T) {
		root := t.TempDir()
		fakeDocker := filepath.Join(root, "docker")
		testutil.WriteFile(t, fakeDocker, "#!/bin/sh\nif [ \"$1\" = ps ]; then printf '\\napp\\timg\\tUp\\n'; fi\n")
		if err := os.Chmod(fakeDocker, 0o755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))

		got := listRunningContainers("docker", root)
		if len(got) != 1 || got[0].ContainerName != "app" {
			t.Fatalf("unexpected containers: %+v", got)
		}
	})
}
