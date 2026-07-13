package service

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/jsonresp"
	"github.com/NobleMajo/explorer-mcp/internal/service/resource"
	"github.com/NobleMajo/explorer-mcp/internal/testutil"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestLimitJSONOutput(t *testing.T) {
	t.Parallel()

	t.Run("under cap", func(t *testing.T) {
		t.Parallel()
		input := `{"entryCount":1}`
		got, err := limitJSONOutput(input)
		if err != nil {
			t.Fatalf("limitJSONOutput() error: %v", err)
		}
		if got != input {
			t.Fatalf("limitJSONOutput() changed small payload")
		}
	})

	t.Run("over cap", func(t *testing.T) {
		t.Parallel()
		input := `{"payload":"` + strings.Repeat("x", maxToolJSONBytes) + `"}`

		got, err := limitJSONOutput(input)
		if err != nil {
			t.Fatalf("limitJSONOutput() error: %v", err)
		}

		var resp truncatedToolResponse
		if err := json.Unmarshal([]byte(got), &resp); err != nil {
			t.Fatalf("unmarshal truncated response: %v", err)
		}

		if !resp.IsOutputTruncated {
			t.Fatal("expected isOutputTruncated true")
		}
		if resp.OutputByteCount <= maxToolJSONBytes {
			t.Fatalf("outputByteCount = %d, want > %d", resp.OutputByteCount, maxToolJSONBytes)
		}
		if len(resp.TruncatedJsonText) != maxToolJSONBytes {
			t.Fatalf("truncatedJsonText len = %d, want %d", len(resp.TruncatedJsonText), maxToolJSONBytes)
		}
	})
}

func TestJsonToolResult(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		result, _, err := jsonToolResult(func() (string, error) {
			return marshalResponse(responseMeta{
				ToolName:      "demo",
				SchemaVersion: jsonresp.SchemaVersion,
			})
		})
		if err != nil {
			t.Fatalf("jsonToolResult() error: %v", err)
		}
		if len(result.Content) != 1 {
			t.Fatalf("content len = %d, want 1", len(result.Content))
		}
	})

	t.Run("propagates error", func(t *testing.T) {
		t.Parallel()
		_, _, err := jsonToolResult(func() (string, error) {
			return "", errors.New("boom")
		})
		if err == nil {
			t.Fatal("expected error propagation")
		}
	})

	t.Run("truncates large payload", func(t *testing.T) {
		t.Parallel()
		large := `{"toolName":"demo","schemaVersion":1,"payload":"` + strings.Repeat("x", maxToolJSONBytes) + `"}`

		result, _, err := jsonToolResult(func() (string, error) {
			return large, nil
		})
		if err != nil {
			t.Fatalf("jsonToolResult() error: %v", err)
		}

		text, ok := result.Content[0].(*mcpsdk.TextContent)
		if !ok {
			t.Fatalf("unexpected content type %T", result.Content[0])
		}

		var resp truncatedToolResponse
		if err := json.Unmarshal([]byte(text.Text), &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if !resp.IsOutputTruncated {
			t.Fatal("expected truncated explore payload")
		}
	})
}

func TestExploreCombinesToolSections(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not in PATH")
	}

	parent := t.TempDir()
	root := filepath.Join(parent, "app")
	sibling := filepath.Join(parent, "other")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(sibling, 0o755); err != nil {
		t.Fatal(err)
	}

	testutil.WriteFile(t, root+"/main.go", "package main\n")
	testutil.WriteFile(t, root+"/go.mod", "module demo\n")
	testutil.WriteFile(t, root+"/Makefile", "build:\n\ntest:\n")
	testutil.WriteFile(t, root+"/Dockerfile", "FROM alpine\n")
	testutil.WriteFile(t, root+"/requirements.txt", "requests==2.28.0\n")

	testutil.Chdir(t, root)

	initCmd := exec.Command("git", "init")
	initCmd.Dir = root
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}

	jsonText, err := buildExploreResponse(testExploreSettings(false))
	if err != nil {
		t.Fatalf("buildExploreResponse() error: %v", err)
	}

	jsonTextVerbose, err := buildExploreResponse(testExploreSettings(true))
	if err != nil {
		t.Fatalf("buildExploreResponse(true) error: %v", err)
	}

	var resp exploreResponse
	testutil.ParseJSON(t, jsonText, &resp)

	if resp.ToolName != "explore" {
		t.Fatalf("toolName = %q", resp.ToolName)
	}
	if resp.ProjectRootPath != root {
		t.Fatalf("projectRootPath = %q, want %q", resp.ProjectRootPath, root)
	}

	assertSectionHasField(t, "structure", resp.Structure, "repoScanDepthLimit")
	assertSectionMissingField(t, "structure", resp.Structure, "rootPath")
	assertSectionMissingField(t, "structure", resp.Structure, "projectRootPath")
	assertSectionHasField(t, "git", resp.Git, "isGitAvailable")
	assertSectionHasField(t, "workspace", resp.Workspace, "parentScanPerformed")
	assertSectionMissingField(t, "workspace", resp.Workspace, "currentWorkingDirectoryPath")
	assertSectionMissingField(t, "workspace", resp.Workspace, "parentDirectoryPath")
	assertSectionIsJSONArray(t, "dependencies", resp.Dependencies)
	assertSectionHasField(t, "container", resp.Container, "cliFound")
	assertSectionMissingField(t, "container", resp.Container, "projectRootPath")
	assertSectionHasField(t, "tools", resp.Tools, "toolsFound")
	assertSectionMissingField(t, "tools", resp.Tools, "projectRootPath")
	assertSectionMissingField(t, "tools", resp.Tools, "hasMakefile")
	assertSectionHasField(t, "cli", resp.CLI, "commonCliToolsFound")

	if resp.AgentBehaviorMainInstruction != AgentBehaviorMainInstruction {
		t.Fatalf("agentBehaviorMainInstruction = %q", resp.AgentBehaviorMainInstruction)
	}

	for _, want := range agentBehaviorInstructionDomains {
		if _, ok := resp.AgentBehaviorInstructions[want]; !ok {
			t.Fatalf("expected agentBehaviorInstructions to include %q, got %v", want, resp.AgentBehaviorInstructions)
		}
		if resp.AgentBehaviorInstructions[want] != AgentBehaviorInstructions[want] {
			t.Fatalf("wrong instruction text for %q", want)
		}
	}

	var respVerbose exploreResponse
	testutil.ParseJSON(t, jsonTextVerbose, &respVerbose)
	if respVerbose.ToolName != "explore" {
		t.Fatalf("verbose toolName = %q", respVerbose.ToolName)
	}
}

func TestBuildExploreResponsePropagatesSectionError(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/package.json", `{invalid`)
	testutil.Chdir(t, root)

	_, err := buildExploreResponse(testExploreSettings(false))
	if err == nil {
		t.Fatal("expected buildExploreResponse error from invalid package.json")
	}
}

func TestBuildExploreResponseMakefileReadError(t *testing.T) {
	root := t.TempDir()
	makefile := filepath.Join(root, "Makefile")
	testutil.WriteFile(t, makefile, "build:\n")
	if err := os.Chmod(makefile, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(makefile, 0o644) })

	testutil.Chdir(t, root)

	_, err := buildExploreResponse(testExploreSettings(false))
	if err == nil {
		t.Fatal("expected buildExploreResponse error from unreadable Makefile")
	}
}

func TestBuildAgentBehaviorInstructionsMinimal(t *testing.T) {
	t.Parallel()

	instructions := buildAgentBehaviorInstructions(exploreSections{
		structure:     mustRawJSON(t, map[string]any{"repoScanDepthLimit": 0}),
		git:       mustRawJSON(t, map[string]any{"isGitRepo": false}),
		workspace:  mustRawJSON(t, map[string]any{"parentScanPerformed": false}),
		dependencies:      mustRawJSON(t, []string{}),
		container: mustRawJSON(t, map[string]any{}),
		tools:      mustRawJSON(t, map[string]any{}),
	})

	if len(instructions) != 0 {
		t.Fatalf("len(instructions) = %d, want 0", len(instructions))
	}
}

func TestShouldIncludeBehaviorHint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		domain string
		sect   exploreSections
		want   bool
	}{
		{
			name:   "structure with entries",
			domain: "structure",
			sect: exploreSections{
				structure: mustRawJSON(t, map[string]any{"repoScanDepthLimit": 6, "entryCount": 2}),
			},
			want: true,
		},
		{
			name:   "structure empty",
			domain: "structure",
			sect: exploreSections{
				structure: mustRawJSON(t, map[string]any{"repoScanDepthLimit": 6, "entryCount": 0}),
			},
			want: false,
		},
		{
			name:   "git repo",
			domain: "git",
			sect: exploreSections{
				git: mustRawJSON(t, map[string]any{"isGitRepo": true}),
			},
			want: true,
		},
		{
			name:   "git not repo",
			domain: "git",
			sect: exploreSections{
				git: mustRawJSON(t, map[string]any{"isGitRepo": false}),
			},
			want: false,
		},
		{
			name:   "parent with sibling",
			domain: "parent",
			sect: exploreSections{
				workspace: mustRawJSON(t, map[string]any{
					"parentScanPerformed": true,
					"siblingProjects":     []string{"../other"},
				}),
			},
			want: true,
		},
		{
			name:   "parent only current",
			domain: "parent",
			sect: exploreSections{
				workspace: mustRawJSON(t, map[string]any{
					"parentScanPerformed": true,
					"siblingProjects":     []string{},
				}),
			},
			want: false,
		},
		{
			name:   "deps with ecosystems",
			domain: "deps",
			sect: exploreSections{
				dependencies: mustRawJSON(t, []string{"demo@1.0.0 @direct"}),
			},
			want: true,
		},
		{
			name:   "deps empty",
			domain: "deps",
			sect: exploreSections{
				dependencies: mustRawJSON(t, []string{}),
			},
			want: false,
		},
		{
			name:   "container files",
			domain: "container",
			sect: exploreSections{
				container: mustRawJSON(t, map[string]any{"cliFound": []string{"docker"}}),
			},
			want: true,
		},
		{
			name:   "container running",
			domain: "container",
			sect: exploreSections{
				container: mustRawJSON(t, map[string]any{"containerFound": map[string]any{"docker": []string{"name@web image@img:1 ports@8080:80 mounts@"}}}),
			},
			want: true,
		},
		{
			name:   "container cli available",
			domain: "container",
			sect: exploreSections{
				container: mustRawJSON(t, map[string]any{"cliFound": []string{"podman"}}),
			},
			want: true,
		},
		{
			name:   "container empty",
			domain: "container",
			sect: exploreSections{
				container: mustRawJSON(t, map[string]any{}),
			},
			want: false,
		},
		{
			name:   "tools makefile",
			domain: "tools",
			sect: exploreSections{
				tools: mustRawJSON(t, map[string]any{"toolsFound": []string{"Makefile"}, "scriptsFound": map[string]any{"make": []string{"build"}}}),
			},
			want: true,
		},
		{
			name:   "tools scripts",
			domain: "tools",
			sect: exploreSections{
				tools: mustRawJSON(t, map[string]any{"toolsFound": []string{"package.json"}, "scriptsFound": map[string]any{"package": []string{"build"}}}),
			},
			want: true,
		},
		{
			name:   "tools empty",
			domain: "tools",
			sect: exploreSections{
				tools: mustRawJSON(t, map[string]any{}),
			},
			want: false,
		},
		{
			name:   "structure invalid json",
			domain: "structure",
			sect: exploreSections{
				structure: json.RawMessage("{"),
			},
			want: false,
		},
		{
			name:   "parent invalid json",
			domain: "parent",
			sect: exploreSections{
				workspace: json.RawMessage("{"),
			},
			want: false,
		},
		{
			name:   "git invalid json",
			domain: "git",
			sect: exploreSections{
				git: json.RawMessage("{"),
			},
			want: false,
		},
		{
			name:   "deps invalid json",
			domain: "deps",
			sect: exploreSections{
				dependencies: json.RawMessage("["),
			},
			want: false,
		},
		{
			name:   "tools shell scripts",
			domain: "tools",
			sect: exploreSections{
				tools: mustRawJSON(t, map[string]any{"toolsFound": []string{"*.sh"}, "scriptsFound": map[string]any{"shell": []string{"run.sh"}}}),
			},
			want: true,
		},
		{
			name:   "unknown domain",
			domain: "unknown",
			sect:   exploreSections{},
			want:   false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := shouldIncludeBehaviorHint(tc.domain, tc.sect); got != tc.want {
				t.Fatalf("shouldIncludeBehaviorHint(%q) = %v, want %v", tc.domain, got, tc.want)
			}
		})
	}
}

func TestHasContainerOverviewData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  json.RawMessage
		want bool
	}{
		{name: "empty", raw: nil, want: false},
		{name: "invalid json", raw: json.RawMessage("{"), want: false},
		{name: "zero counts", raw: mustRawJSON(t, map[string]any{"cliFound": []string{}, "containerFound": map[string]any{}}), want: false},
		{name: "cli", raw: mustRawJSON(t, map[string]any{"cliFound": []string{"docker"}}), want: true},
		{name: "running", raw: mustRawJSON(t, map[string]any{"containerFound": map[string]any{"docker": []string{"name@web"}}}), want: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := hasContainerOverviewData(tc.raw); got != tc.want {
				t.Fatalf("hasContainerOverviewData() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHasProjectToolsData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  json.RawMessage
		want bool
	}{
		{name: "empty", raw: nil, want: false},
		{name: "invalid json", raw: json.RawMessage("{"), want: false},
		{name: "zero counts", raw: mustRawJSON(t, map[string]any{"toolsFound": []string{}, "scriptsFound": map[string]any{}}), want: false},
		{name: "makefile", raw: mustRawJSON(t, map[string]any{"toolsFound": []string{"Makefile"}}), want: true},
		{name: "package json", raw: mustRawJSON(t, map[string]any{"scriptsFound": map[string]any{"package": []string{"build"}}}), want: true},
		{name: "shell", raw: mustRawJSON(t, map[string]any{"scriptsFound": map[string]any{"shell": []string{"run.sh"}}}), want: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := hasProjectToolsData(tc.raw); got != tc.want {
				t.Fatalf("hasProjectToolsData() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBuildAgentBehaviorInstructions(t *testing.T) {
	t.Parallel()

	sections := exploreSections{
		structure:     mustRawJSON(t, map[string]any{"repoScanDepthLimit": 6, "entryCount": 1}),
		git:       mustRawJSON(t, map[string]any{"isGitRepo": true}),
		workspace:  mustRawJSON(t, map[string]any{"parentScanPerformed": true, "siblingProjects": []string{"../other"}}),
		dependencies:      mustRawJSON(t, []string{"demo@1.0.0 @direct"}),
		container: mustRawJSON(t, map[string]any{"cliFound": []string{"docker"}}),
		tools:      mustRawJSON(t, map[string]any{"toolsFound": []string{"Makefile"}, "scriptsFound": map[string]any{"make": []string{"build"}}}),
	}

	instructions := buildAgentBehaviorInstructions(sections)

	for _, domain := range agentBehaviorInstructionDomains {
		if instructions[domain] != AgentBehaviorInstructions[domain] {
			t.Fatalf("missing or wrong instruction for %q", domain)
		}
	}
	if len(instructions) != len(agentBehaviorInstructionDomains) {
		t.Fatalf("len(instructions) = %d, want %d", len(instructions), len(agentBehaviorInstructionDomains))
	}
}

func TestBuildAgentBehaviorInstructionsSkipsEmptyDomainText(t *testing.T) {
	t.Parallel()

	catalog := make(map[string]string, len(AgentBehaviorInstructions))
	for key, value := range AgentBehaviorInstructions {
		catalog[key] = value
	}
	catalog["git"] = ""

	sections := exploreSections{
		structure: mustRawJSON(t, map[string]any{"repoScanDepthLimit": 6, "entryCount": 0}),
		git:   mustRawJSON(t, map[string]any{"isGitRepo": true}),
	}

	instructions := buildAgentBehaviorInstructionsWith(sections, catalog)
	if _, ok := instructions["git"]; ok {
		t.Fatal("expected git instruction to be skipped when text is empty")
	}
}

func TestOverviewSection(t *testing.T) {
	t.Parallel()

	t.Run("embeds json", func(t *testing.T) {
		t.Parallel()
		section, err := overviewSection(func() resource.OverviewResource {
			return func(verbose bool) (any, error) {
				if verbose {
					t.Fatal("expected verbose false")
				}
				return map[string]int{"entryCount": 1}, nil
			}
		}, false)
		if err != nil {
			t.Fatalf("overviewSection() error: %v", err)
		}

		var sectionData map[string]int
		if err := json.Unmarshal(section, &sectionData); err != nil {
			t.Fatalf("unmarshal section: %v", err)
		}
		if sectionData["entryCount"] != 1 {
			t.Fatalf("entryCount = %d", sectionData["entryCount"])
		}
	})

	t.Run("passes verbose", func(t *testing.T) {
		t.Parallel()
		section, err := overviewSection(func() resource.OverviewResource {
			return func(verbose bool) (any, error) {
				if !verbose {
					t.Fatal("expected verbose true")
				}
				return map[string]bool{"verbose": true}, nil
			}
		}, true)
		if err != nil {
			t.Fatalf("overviewSection() error: %v", err)
		}

		var sectionData map[string]bool
		if err := json.Unmarshal(section, &sectionData); err != nil {
			t.Fatalf("unmarshal section: %v", err)
		}
		if !sectionData["verbose"] {
			t.Fatal("expected verbose true in section payload")
		}
	})

	t.Run("propagates error", func(t *testing.T) {
		t.Parallel()
		_, err := overviewSection(func() resource.OverviewResource {
			return func(verbose bool) (any, error) {
				_ = verbose
				return nil, errors.New("section failed")
			}
		}, false)
		if err == nil {
			t.Fatal("expected overviewSection error")
		}
	})

	t.Run("marshal error", func(t *testing.T) {
		t.Parallel()
		_, err := overviewSection(func() resource.OverviewResource {
			return func(verbose bool) (any, error) {
				_ = verbose
				return make(chan int), nil
			}
		}, false)
		if err == nil {
			t.Fatal("expected marshal error")
		}
	})
}

func mustRawJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return data
}

func assertSectionIsJSONArray(t *testing.T, field string, raw json.RawMessage) {
	t.Helper()

	var section []json.RawMessage
	if err := json.Unmarshal(raw, &section); err != nil {
		t.Fatalf("unmarshal %s as array: %v", field, err)
	}
}

func assertSectionMissingField(t *testing.T, field string, raw json.RawMessage, unwantedField string) {
	t.Helper()

	var section map[string]json.RawMessage
	if err := json.Unmarshal(raw, &section); err != nil {
		t.Fatalf("unmarshal %s: %v", field, err)
	}
	if _, ok := section[unwantedField]; ok {
		t.Fatalf("%s should not include field %q", field, unwantedField)
	}
}

func assertSectionHasField(t *testing.T, field string, raw json.RawMessage, wantField string) {
	t.Helper()

	var section map[string]json.RawMessage
	if err := json.Unmarshal(raw, &section); err != nil {
		t.Fatalf("unmarshal %s: %v", field, err)
	}
	if _, ok := section[wantField]; !ok {
		t.Fatalf("%s missing field %q", field, wantField)
	}
}

func TestBuildExploreResponseDisabledScansOmitArrays(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, root+"/main.go", "package main\n")
	testutil.Chdir(t, root)

	jsonText, err := buildExploreResponse(exploreSettings{
		recentCommitCount: 0,
		parentScanDepth:   0,
		repoScanDepth:     0,
	})
	if err != nil {
		t.Fatalf("buildExploreResponse() error: %v", err)
	}

	var resp exploreResponse
	testutil.ParseJSON(t, jsonText, &resp)

	assertSectionHasField(t, "structure", resp.Structure, "repoScanDepthLimit")
	assertSectionMissingField(t, "structure", resp.Structure, "repoScanPerformed")
	assertSectionMissingField(t, "structure", resp.Structure, "entries")
	assertSectionMissingField(t, "structure", resp.Structure, "entryCount")

	assertSectionHasField(t, "git", resp.Git, "recentCommitsListed")
	assertSectionMissingField(t, "git", resp.Git, "recentCommitCount")
	assertSectionMissingField(t, "git", resp.Git, "someRecentCommits")

	assertSectionHasField(t, "workspace", resp.Workspace, "parentScanPerformed")
	assertSectionMissingField(t, "workspace", resp.Workspace, "siblingProjects")
	assertSectionMissingField(t, "workspace", resp.Workspace, "siblingProjectCount")
}

func testExploreSettings(verbose bool) exploreSettings {
	return exploreSettings{
		verbose:           verbose,
		recentCommitCount: 10,
		parentScanDepth:   3,
		repoScanDepth:     7,
	}
}
