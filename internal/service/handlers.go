package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/NobleMajo/explorer-mcp/internal/jsonresp"
	"github.com/NobleMajo/explorer-mcp/internal/service/overviews/cli"
	"github.com/NobleMajo/explorer-mcp/internal/service/overviews/container"
	"github.com/NobleMajo/explorer-mcp/internal/service/overviews/deps"
	"github.com/NobleMajo/explorer-mcp/internal/service/overviews/git"
	"github.com/NobleMajo/explorer-mcp/internal/service/overviews/opencode"
	"github.com/NobleMajo/explorer-mcp/internal/service/overviews/parent"
	"github.com/NobleMajo/explorer-mcp/internal/service/overviews/structure"
	"github.com/NobleMajo/explorer-mcp/internal/service/overviews/tools"
	"github.com/NobleMajo/explorer-mcp/internal/service/resource"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const maxToolJSONBytes = 32 * 1024

type responseMeta = jsonresp.Meta

func marshalResponse(v any) (string, error) {
	return jsonresp.Marshal(v)
}

var readOnlyToolAnnotations = &mcpsdk.ToolAnnotations{
	ReadOnlyHint: true,
}

func registerExploreTool(server *mcpsdk.Server, settings exploreSettings) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "explore",
		Description: "Workspace overview as JSON with structure, git, workspace, dependencies, container, tools, cli, opencode, agentBehaviorMainInstruction, and agentBehaviorInstructions",
		Annotations: readOnlyToolAnnotations,
	}, func(ctx context.Context, req *mcpsdk.CallToolRequest, _ struct{}) (*mcpsdk.CallToolResult, any, error) {
		start := time.Now()
		if settings.verbose {
			_, _ = fmt.Fprintf(os.Stderr, "explore: request begin\n")
		}

		return jsonToolResult(func() (string, error) {
			text, err := buildExploreResponse(settings)
			if settings.verbose {
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "explore: request end status=error dur=%s\n", time.Since(start))
				} else {
					_, _ = fmt.Fprintf(os.Stderr, "explore: request end status=ok dur=%s\n", time.Since(start))
				}
			}
			return text, err
		})
	})
}

var AgentBehaviorMainInstruction = "You must strictly follow the specific behavior guidelines below. Always call this MCP endpoint when preparing the next development step, or whenever the project state, files, or execution context could have changed."
var AgentBehaviorInstructions = map[string]string{
	"container": "Do not restart or stop discovered containers if they mount local source code and use auto-restart policies. Analyze container execution via runtime logs, local configurations, and container metadata. Actively scan for: compose.yml, compose.yaml, docker-compose.yml, docker-compose.yaml, Dockerfile, *.dockerfile, the ./docker directory, and related runtime assets.",
	"cli":       "Prefer using existing CLI tools found in this environment over adding dependencies or writing ad-hoc scripts. When proposing commands, keep them minimal, reproducible, and aligned with available tooling.",
	"deps":      "Minimize dependencies. Avoid adding unused or redundant packages. Focus strictly on the target requirement and prefer native standard libraries where applicable. Locate dependency source code paths and verify if a specialized docs-mcp exists to analyze and predict external package behavior before making structural modifications.",
	"git":       "When requested to recommend commits or perform a commit operation, always group uncommitted changes into distinct, logical atomic commits. For each proposed commit, provide exactly 3 structured commit message variants and a clear description of the specific changes. NEVER execute a commit autonomously unless explicitly and directly instructed to do so.",
	"opencode":  "If OpenCode permissions are present, give preference to 'Allow' over 'Ask'. Always order commands, subcommands, flags and arguments in the permitted way. If a prompt relates to an existing MCP server, use read-only operations such as 'docs' and 'explore' to gain more context before starting the task.",
	"parent":    "Scan parent and sibling directories to identify external dependencies, architectural services, cross-project tools, or related microservices that reside adjacent to the current workspace root directory.",
	"structure": "Strict adherence to the established project layout is mandatory. Do not alter the directory schema unless the requested feature explicitly forces a deviation from current file and folder patterns. Analyze the existing codebase layout to derive and follow local structural conventions and architectural design patterns. Entries ending in /** mark directories that contain additional files or subdirectories below repoScanDepthLimit; treat them as proof that deeper layout exists even though those paths are not listed.",
	"tools":     "Analyze and inventory existing tooling configurations for testing, linting, building, executing, and container orchestration. Prioritize utilizing predefined Makefile targets, local scripts, and existing automation tools over generating new standalone commands or chaining raw shell operations.",
}

type exploreResponse struct {
	responseMeta
	ProjectRootPath              string            `json:"projectRootPath"`
	Structure                    json.RawMessage   `json:"structure,omitempty"`
	Git                          json.RawMessage   `json:"git,omitempty"`
	Workspace                    json.RawMessage   `json:"workspace,omitempty"`
	Dependencies                 json.RawMessage   `json:"dependencies,omitempty"`
	Container                    json.RawMessage   `json:"container,omitempty"`
	Tools                        json.RawMessage   `json:"tools,omitempty"`
	CLI                          json.RawMessage   `json:"cli,omitempty"`
	Opencode                     json.RawMessage   `json:"opencode,omitempty"`
	AgentBehaviorMainInstruction string            `json:"agentBehaviorMainInstruction,omitempty"`
	AgentBehaviorInstructions    map[string]string `json:"agentBehaviorInstructions,omitempty"`
}

func buildExploreResponse(settings exploreSettings) (string, error) {
	if !settings.hasEnabledOverview() {
		return "", ErrAllOverviewsDisabled
	}

	logf := func(format string, args ...any) {
		if !settings.verbose {
			return
		}
		_, _ = fmt.Fprintf(os.Stderr, "explore: "+format+"\n", args...)
	}

	projectRoot, err := os.Getwd()
	if err != nil {
		logf("getwd failed: %v", err)
		return "", err
	}

	runSection := func(name string, disabled bool, fn func() resource.OverviewResource) (json.RawMessage, error) {
		if disabled {
			logf("%s: begin status=disabled", name)
			return nil, nil
		}
		logf("%s: begin", name)
		start := time.Now()
		section, err := overviewSection(fn, settings.verbose)
		dur := time.Since(start)
		if err != nil {
			logf("%s: end status=error dur=%s", name, dur)
			return nil, err
		}
		if section == nil {
			logf("%s: end status=omitted dur=%s", name, dur)
			return nil, nil
		}
		logf("%s: end status=included dur=%s", name, dur)
		return section, nil
	}

	repoStructure, err := runSection("structure", settings.disableStructureOverview, structure.StructureOverview(settings.repoScanDepth))
	if err != nil {
		return "", err
	}

	gitOverview, err := runSection("git", settings.disableGitOverview, git.GitOverview(settings.recentCommitCount))
	if err != nil {
		return "", err
	}

	workspaceContext, err := runSection("workspace", settings.disableWorkspaceOverview, parent.ParentOverview(settings.parentScanSettings()))
	if err != nil {
		return "", err
	}

	dependencies, err := runSection("dependencies", settings.disableDependenciesOverview, deps.DepsOverview)
	if err != nil {
		return "", err
	}

	containerOverview, err := runSection("container", settings.disableContainerOverview, container.ContainerOverview)
	if err != nil {
		return "", err
	}

	projectTools, err := runSection("tools", settings.disableToolsOverview, tools.ToolsOverview)
	if err != nil {
		return "", err
	}

	cliOverview, err := runSection("cli", !settings.enableCliOverview, cli.CLIOverview)
	if err != nil {
		return "", err
	}

	opencodeOverview, err := runSection("opencode", !settings.enableOpencodeOverview, opencode.OpencodeOverview)
	if err != nil {
		return "", err
	}

	sections := exploreSections{
		structure:    repoStructure,
		git:          gitOverview,
		workspace:    workspaceContext,
		dependencies: dependencies,
		container:    containerOverview,
		tools:        projectTools,
		cli:          cliOverview,
		opencode:     opencodeOverview,
	}

	response := exploreResponse{
		responseMeta: responseMeta{
			ToolName:      "explore",
			SchemaVersion: jsonresp.SchemaVersion,
		},
		ProjectRootPath: projectRoot,
		Structure:       sections.structure,
		Git:             sections.git,
		Workspace:       sections.workspace,
		Dependencies:    sections.dependencies,
		Container:       sections.container,
		Tools:           sections.tools,
		CLI:             sections.cli,
		Opencode:        sections.opencode,
	}

	if settings.enableBehaviorInstruction {
		response.AgentBehaviorMainInstruction = AgentBehaviorMainInstruction
		response.AgentBehaviorInstructions = buildAgentBehaviorInstructions(sections)
	}

	return marshalResponse(response)
}

var agentBehaviorInstructionDomains = []string{
	"structure",
	"git",
	"parent",
	"deps",
	"container",
	"tools",
	"cli",
	"opencode",
}

func buildAgentBehaviorInstructions(sections exploreSections) map[string]string {
	return buildAgentBehaviorInstructionsWith(sections, AgentBehaviorInstructions)
}

func buildAgentBehaviorInstructionsWith(sections exploreSections, catalog map[string]string) map[string]string {
	instructions := make(map[string]string)

	for _, domainName := range agentBehaviorInstructionDomains {
		if !shouldIncludeBehaviorHint(domainName, sections) {
			continue
		}
		text, ok := catalog[domainName]
		if !ok || text == "" {
			continue
		}
		instructions[domainName] = text
	}

	return instructions
}

type exploreSections struct {
	structure    json.RawMessage
	git          json.RawMessage
	workspace    json.RawMessage
	dependencies json.RawMessage
	container    json.RawMessage
	tools        json.RawMessage
	cli          json.RawMessage
	opencode     json.RawMessage
}

func shouldIncludeBehaviorHint(domainName string, sections exploreSections) bool {
	switch domainName {
	case "structure":
		var structure struct {
			RepoScanDepthLimit int `json:"repoScanDepthLimit"`
			EntryCount         int `json:"entryCount"`
		}
		if json.Unmarshal(sections.structure, &structure) != nil {
			return false
		}
		if structure.RepoScanDepthLimit < 1 {
			return false
		}
		return structure.EntryCount > 0
	case "git":
		var git struct {
			IsGitRepo bool `json:"isGitRepo"`
		}
		return json.Unmarshal(sections.git, &git) == nil && git.IsGitRepo
	case "parent":
		var parent struct {
			ParentScanPerformed bool     `json:"parentScanPerformed"`
			SiblingProjects     []string `json:"siblingProjects"`
		}
		if json.Unmarshal(sections.workspace, &parent) != nil {
			return false
		}
		if !parent.ParentScanPerformed {
			return false
		}
		return len(parent.SiblingProjects) > 0
	case "deps":
		var deps []string
		return json.Unmarshal(sections.dependencies, &deps) == nil && len(deps) > 0
	case "container":
		return hasContainerOverviewData(sections.container)
	case "tools":
		return hasProjectToolsData(sections.tools)
	case "cli":
		return len(sections.cli) > 0
	case "opencode":
		return len(sections.opencode) > 0
	default:
		return false
	}
}

func hasContainerOverviewData(containerOverview json.RawMessage) bool {
	if len(containerOverview) == 0 {
		return false
	}

	var overview struct {
		CLIFound       []string            `json:"cliFound"`
		ContainerFound map[string][]string `json:"containerFound"`
	}
	if json.Unmarshal(containerOverview, &overview) != nil {
		return false
	}

	if len(overview.CLIFound) > 0 {
		return true
	}
	for _, containers := range overview.ContainerFound {
		if len(containers) > 0 {
			return true
		}
	}
	return false
}

func hasProjectToolsData(projectTools json.RawMessage) bool {
	if len(projectTools) == 0 {
		return false
	}

	var tools struct {
		ToolsFound   []string            `json:"toolsFound"`
		ScriptsFound map[string][]string `json:"scriptsFound"`
	}
	if json.Unmarshal(projectTools, &tools) != nil {
		return false
	}

	if len(tools.ToolsFound) > 0 {
		return true
	}
	for _, scripts := range tools.ScriptsFound {
		if len(scripts) > 0 {
			return true
		}
	}
	return false
}

func optionalOverviewSection(disabled bool, fn func() resource.OverviewResource, verbose bool) (json.RawMessage, error) {
	if disabled {
		return nil, nil
	}
	return overviewSection(fn, verbose)
}

func overviewSection(fn func() resource.OverviewResource, verbose bool) (json.RawMessage, error) {
	result, err := fn()(verbose)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	text, err := marshalResponse(result)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(text), nil
}

type truncatedToolResponse struct {
	responseMeta
	IsOutputTruncated bool   `json:"isOutputTruncated"`
	OutputByteCount   int    `json:"outputByteCount"`
	TruncatedJsonText string `json:"truncatedJsonText"`
}

func limitJSONOutput(jsonText string) (string, error) {
	if len(jsonText) <= maxToolJSONBytes {
		return jsonText, nil
	}

	var meta struct {
		ToolName string `json:"toolName"`
	}
	_ = json.Unmarshal([]byte(jsonText), &meta)

	return marshalResponse(truncatedToolResponse{
		responseMeta: responseMeta{
			ToolName:      meta.ToolName,
			SchemaVersion: jsonresp.SchemaVersion,
		},
		IsOutputTruncated: true,
		OutputByteCount:   len(jsonText),
		TruncatedJsonText: jsonText[:maxToolJSONBytes],
	})
}

func jsonToolResult(fn func() (string, error)) (*mcpsdk.CallToolResult, any, error) {
	jsonText, err := fn()
	if err != nil {
		return nil, nil, err
	}

	jsonText, err = limitJSONOutput(jsonText)
	if err != nil {
		return nil, nil, err
	}

	return &mcpsdk.CallToolResult{
		Content: []mcpsdk.Content{
			&mcpsdk.TextContent{Text: jsonText},
		},
	}, nil, nil
}
