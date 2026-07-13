package tools

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestParseMakefileTargetNames(t *testing.T) {
	t.Parallel()

	content := "# comment\n.PHONY: build\nbuild test: deps\n\tignored-line\n.dockerignore: x\ninstall:\n"
	got := parseMakefileTargetNames(content)
	want := []string{"build", "install", "test"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseMakefileTargetNames() = %v, want %v", got, want)
	}
}

func TestParseMakefileTargetNamesDedupesTargets(t *testing.T) {
	t.Parallel()

	got := parseMakefileTargetNames("build build: deps\n")
	if len(got) != 1 || got[0] != "build" {
		t.Fatalf("parseMakefileTargetNames() = %v, want [build]", got)
	}
}

func TestParsePackageJsonScriptNames(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "package.json")
	testutil.WriteFile(t, path, `{"scripts":{"build":"tsc","test":"vitest"}}`)

	got, err := parsePackageJsonScriptNames(path)
	if err != nil {
		t.Fatalf("parsePackageJsonScriptNames() error: %v", err)
	}
	want := []string{"build", "test"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parsePackageJsonScriptNames() = %v, want %v", got, want)
	}
}

func TestParsePackageJsonScriptNamesInvalidJSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "package.json")
	testutil.WriteFile(t, path, `{invalid`)

	_, err := parsePackageJsonScriptNames(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestListRootShellScripts(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "run.sh"), "#!/bin/sh\n")
	testutil.WriteFile(t, filepath.Join(root, "README.md"), "x\n")
	if err := os.MkdirAll(filepath.Join(root, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, filepath.Join(root, "scripts", "nested.sh"), "#!/bin/sh\n")

	got, err := listRootShellScripts(root)
	if err != nil {
		t.Fatalf("listRootShellScripts() error: %v", err)
	}
	if !reflect.DeepEqual(got, []string{"run.sh"}) {
		t.Fatalf("listRootShellScripts() = %v, want [run.sh]", got)
	}
}

func TestListRootShellScriptsMissingDir(t *testing.T) {
	t.Parallel()

	_, err := listRootShellScripts("/does/not/exist")
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
}

func TestParsePackageJsonScriptNamesMissingFile(t *testing.T) {
	t.Parallel()

	_, err := parsePackageJsonScriptNames(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParseMakefileTargetNamesSkipsAssignmentsAndShellLines(t *testing.T) {
	t.Parallel()

	content := `PROJECT_VERSION := 1.0.0
PORT ?= 8080
help: ##@ prints help
build test: deps
	  echo "- os: linux";
@echo recipe
.PHONY: docker
docker: ##@ run container
`
	got := parseMakefileTargetNames(content)
	want := []string{"build", "docker", "help", "test"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseMakefileTargetNames() = %v, want %v", got, want)
	}
}

func TestParseMakefileTargetNamesSkipsNonTargets(t *testing.T) {
	t.Parallel()

	got := parseMakefileTargetNames("help:\n\t@echo ok\nnot-a-target\n")
	if len(got) != 1 || got[0] != "help" {
		t.Fatalf("parseMakefileTargetNames() = %v, want [help]", got)
	}
}

func TestProjectToolsMakefileReadError(t *testing.T) {
	root := t.TempDir()
	makefile := filepath.Join(root, "Makefile")
	testutil.WriteFile(t, makefile, "build:\n")
	if err := os.Chmod(makefile, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(makefile, 0o644) })
	testutil.Chdir(t, root)

	_, err := ToolsOverview()(false)
	if err == nil {
		t.Fatal("expected error for unreadable Makefile")
	}
}

func TestProjectToolsDetectsPackageJsonScripts(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "package.json"), `{"scripts":{"build":"tsc","test":"vitest"}}`)

	testutil.Chdir(t, root)

	result, err := ToolsOverview()(false)
	if err != nil {
		t.Fatalf("ToolsOverview() error: %v", err)
	}

	resp, ok := result.(projectToolsResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	if !resp.HasPackageJson || resp.PackageJsonScriptCount != 2 {
		t.Fatalf("unexpected package.json scripts: %+v", resp)
	}
}

func TestProjectToolsDetectsMakefileTargets(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "Makefile"), "build:\n\ntest:\n")
	testutil.WriteFile(t, filepath.Join(root, "run.sh"), "#!/bin/sh\n")

	testutil.Chdir(t, root)

	result, err := ToolsOverview()(false)
	if err != nil {
		t.Fatalf("ToolsOverview() error: %v", err)
	}

	resp, ok := result.(projectToolsResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	if !resp.HasMakefile || resp.MakefileTargetCount != 2 {
		t.Fatalf("unexpected makefile targets: %+v", resp)
	}
	if resp.ShellScriptCount != 1 {
		t.Fatalf("shellScriptCount = %d, want 1", resp.ShellScriptCount)
	}
}
