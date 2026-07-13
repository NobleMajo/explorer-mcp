package container

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestDetectAvailableContainerCLIs(t *testing.T) {
	binDir := t.TempDir()
	for _, name := range []string{"docker", "buildah", "skopeo"} {
		path := filepath.Join(binDir, name)
		testutil.WriteFile(t, path, "#!/bin/sh\n")
		if err := os.Chmod(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", binDir)

	got := detectAvailableContainerCLIs()
	want := []string{"docker", "buildah", "skopeo"}
	if !slices.Equal(got, want) {
		t.Fatalf("detectAvailableContainerCLIs() = %v, want %v", got, want)
	}
}

func TestContainerOverviewIntegration(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "Dockerfile"), "FROM alpine\n")
	testutil.Chdir(t, root)

	binDir := t.TempDir()
	for _, runtime := range []string{"docker", "podman"} {
		script := filepath.Join(binDir, runtime)
		testutil.WriteFile(t, script, `#!/bin/sh
if [ "$1" = ps ]; then
  printf 'web\timg:1\t8080:80\n'
  exit 0
fi
if [ "$1" = inspect ]; then
  printf '/host:/app\n'
  exit 0
fi
exit 1
`)
		if err := os.Chmod(script, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", binDir)

	result, err := ContainerOverview()(root, false)
	if err != nil {
		t.Fatalf("ContainerOverview() error: %v", err)
	}

	resp, ok := result.(containerOverviewResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	if !slices.Equal(resp.CLIFound, []string{"docker", "podman"}) {
		t.Fatalf("cliFound = %v", resp.CLIFound)
	}
	if len(resp.ContainerFound["docker"]) != 1 {
		t.Fatalf("docker containers = %v", resp.ContainerFound["docker"])
	}
	if len(resp.ContainerFound["podman"]) != 1 {
		t.Fatalf("podman containers = %v", resp.ContainerFound["podman"])
	}
	for runtime, entries := range resp.ContainerFound {
		if !strings.Contains(entries[0], "name@web image@img:1 ports@8080:80 mounts@/host:/app") {
			t.Fatalf("%s entry = %q", runtime, entries[0])
		}
	}
}

func TestListRunningContainerEntries(t *testing.T) {
	t.Run("parses output", func(t *testing.T) {
		root := t.TempDir()
		fakeDocker := filepath.Join(root, "docker")
		testutil.WriteFile(t, fakeDocker, `#!/bin/sh
if [ "$1" = ps ]; then
  printf 'app\tnginx:alpine\t0.0.0.0:8080->80/tcp\n'
  printf 'db\tpostgres:15\t\n'
  exit 0
fi
if [ "$1" = inspect ]; then
  printf '/data:/var/lib/postgresql/data\n'
  exit 0
fi
`)
		if err := os.Chmod(fakeDocker, 0o755); err != nil {
			t.Fatal(err)
		}

		t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))

		got := listRunningContainerEntries("docker", root)
		if len(got) != 2 {
			t.Fatalf("len(entries) = %d, want 2, got %+v", len(got), got)
		}
		if !strings.Contains(got[0], "name@app image@nginx:alpine ports@0.0.0.0:8080->80/tcp") {
			t.Fatalf("unexpected first entry: %q", got[0])
		}
		if !strings.Contains(got[1], "name@db image@postgres:15") {
			t.Fatalf("unexpected second entry: %q", got[1])
		}
	})

	t.Run("exec error", func(t *testing.T) {
		got := listRunningContainerEntries("definitely-not-a-runtime-cmd", t.TempDir())
		if len(got) != 0 {
			t.Fatalf("expected empty slice on exec error, got %#v", got)
		}
	})

	t.Run("empty line skipped", func(t *testing.T) {
		root := t.TempDir()
		fakeDocker := filepath.Join(root, "docker")
		testutil.WriteFile(t, fakeDocker, `#!/bin/sh
if [ "$1" = ps ]; then printf '\napp\timg\t8080:80\n'; fi
if [ "$1" = inspect ]; then printf '/host:/app\n'; fi
`)
		if err := os.Chmod(fakeDocker, 0o755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))

		got := listRunningContainerEntries("docker", root)
		if len(got) != 1 || !strings.Contains(got[0], "name@app") {
			t.Fatalf("unexpected entries: %+v", got)
		}
	})
}

func TestContainerOverviewCLIsOnlyWhenNothingRunning(t *testing.T) {
	root := t.TempDir()
	testutil.Chdir(t, root)

	binDir := t.TempDir()
	script := filepath.Join(binDir, "docker")
	testutil.WriteFile(t, script, "#!/bin/sh\nexit 1\n")
	if err := os.Chmod(script, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)

	result, err := ContainerOverview()(root, false)
	if err != nil {
		t.Fatalf("ContainerOverview() error: %v", err)
	}

	resp := result.(containerOverviewResponse)
	if !slices.Equal(resp.CLIFound, []string{"docker"}) {
		t.Fatalf("cliFound = %v", resp.CLIFound)
	}
	if resp.ContainerFound != nil && len(resp.ContainerFound) != 0 {
		t.Fatalf("expected no containerFound, got %v", resp.ContainerFound)
	}
}

func TestContainerOverviewEmptyWhenNoCLIs(t *testing.T) {
	root := t.TempDir()
	testutil.Chdir(t, root)
	t.Setenv("PATH", t.TempDir())

	result, err := ContainerOverview()(root, false)
	if err != nil {
		t.Fatalf("ContainerOverview() error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil when no container CLIs in PATH, got %+v", result)
	}
}

func TestContainerMountsUsesContainerNameArg(t *testing.T) {
	root := t.TempDir()
	fakeDocker := filepath.Join(root, "docker")
	testutil.WriteFile(t, fakeDocker, `#!/bin/sh
if [ "$1" = inspect ]; then
  printf 'mount-for-%s\n' "$4"
  exit 0
fi
exit 1
`)
	if err := os.Chmod(fakeDocker, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", root)

	got := containerMounts("docker", root, "api")
	if got != "mount-for-api" {
		t.Fatalf("containerMounts() = %q, want mount-for-api", got)
	}
}

func TestFormatContainerEntry(t *testing.T) {
	t.Parallel()

	got := formatContainerEntry("web", "img:1", "8080:80", "/host:/app")
	want := "name@web image@img:1 ports@8080:80 mounts@/host:/app"
	if got != want {
		t.Fatalf("formatContainerEntry() = %q, want %q", got, want)
	}
}
