package service

import (
	"context"
	"encoding/json"
	"os"

	"github.com/NobleMajo/explorer-mcp/internal/jsonresp"
	"github.com/NobleMajo/explorer-mcp/internal/service/overviews/container"
	"github.com/NobleMajo/explorer-mcp/internal/service/overviews/deps"
	"github.com/NobleMajo/explorer-mcp/internal/service/overviews/git"
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

func registerExploreTool(server *mcpsdk.Server, verbose bool) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "explore",
		Description: "Workspace overview as JSON with repoStructure, gitOverview, workspaceContext, dependencies, containerOverview, projectTools, agentBehaviorMainInstruction, and agentBehaviorInstructions",
		Annotations: readOnlyToolAnnotations,
	}, func(ctx context.Context, req *mcpsdk.CallToolRequest, _ struct{}) (*mcpsdk.CallToolResult, any, error) {
		return jsonToolResult(func() (string, error) {
			return buildExploreResponse(verbose)
		})
	})
}

var AgentBehaviorMainInstruction = "You must strictly adhere to the specific behavior guidelines below whenever their corresponding metadata keys are present in the explorer payload. Always call this MCP endpoint when preparing the next development step, or whenever the project state, files, or execution context could have changed."
var AgentBehaviorInstructions = map[string]string{
	"container": "Do not restart or stop discovered containers if they mount local source code and use auto-restart policies. Analyze container execution via runtime logs, local configurations, and container metadata. Actively scan for: compose.yml, compose.yaml, docker-compose.yml, docker-compose.yaml, Dockerfile, *.dockerfile, the ./docker directory, and related runtime assets.",
	"deps":      "Minimize dependencies. Avoid adding unused or redundant packages. Focus strictly on the target requirement and prefer native standard libraries where applicable. Locate dependency source code paths and verify if a specialized docs-mcp exists to analyze and predict external package behavior before making structural modifications.",
	"git":       "When requested to recommend commits or perform a commit operation, always group uncommitted changes into distinct, logical atomic commits. For each proposed commit, provide exactly 3 structured commit message variants and a clear description of the specific changes. NEVER execute a commit autonomously unless explicitly and directly instructed to do so.",
	"parent":    "Scan parent and sibling directories to identify external dependencies, architectural services, cross-project tools, or related microservices that reside adjacent to the current workspace root directory.",
	"structure": "Strict adherence to the established project layout is mandatory. Do not alter the directory schema unless the requested feature explicitly forces a deviation from current file and folder patterns. Analyze the existing codebase layout to derive and follow local structural conventions and architectural design patterns.",
	"tools":     "Analyze and inventory existing tooling configurations for testing, linting, building, executing, and container orchestration. Prioritize utilizing predefined Makefile targets, local scripts, and existing automation tools over generating new standalone commands or chaining raw shell operations.",
}

type exploreResponse struct {
	responseMeta
	ProjectRootPath     string            `json:"projectRootPath"`
	RepoStructure       json.RawMessage   `json:"repoStructure"`
	GitOverview         json.RawMessage   `json:"gitOverview"`
	WorkspaceContext    json.RawMessage   `json:"workspaceContext"`
	Dependencies        json.RawMessage   `json:"dependencies"`
	ContainerOverview   json.RawMessage   `json:"containerOverview"`
	ProjectTools                   json.RawMessage   `json:"projectTools"`
	AgentBehaviorMainInstruction   string            `json:"agentBehaviorMainInstruction"`
	AgentBehaviorInstructions      map[string]string `json:"agentBehaviorInstructions"`
}

func buildExploreResponse(verbose bool) (string, error) {
	projectRoot, err := os.Getwd()
	if err != nil {
		return "", err
	}

	repoStructure, err := overviewSection(structure.StructureOverview, verbose)
	if err != nil {
		return "", err
	}

	gitOverview, err := overviewSection(git.GitOverview, verbose)
	if err != nil {
		return "", err
	}

	workspaceContext, err := overviewSection(parent.ParentOverview, verbose)
	if err != nil {
		return "", err
	}

	dependencies, err := overviewSection(deps.DepsOverview, verbose)
	if err != nil {
		return "", err
	}

	containerOverview, err := overviewSection(container.ContainerOverview, verbose)
	if err != nil {
		return "", err
	}

	projectTools, err := overviewSection(tools.ToolsOverview, verbose)
	if err != nil {
		return "", err
	}

	sections := exploreSections{
		repoStructure:     repoStructure,
		gitOverview:       gitOverview,
		workspaceContext:  workspaceContext,
		dependencies:      dependencies,
		containerOverview: containerOverview,
		projectTools:      projectTools,
	}

	return marshalResponse(exploreResponse{
		responseMeta: responseMeta{
			ToolName:      "explore",
			SchemaVersion: jsonresp.SchemaVersion,
		},
		ProjectRootPath:     projectRoot,
		RepoStructure:       sections.repoStructure,
		GitOverview:         sections.gitOverview,
		WorkspaceContext:    sections.workspaceContext,
		Dependencies:        sections.dependencies,
		ContainerOverview:   sections.containerOverview,
		ProjectTools:                 sections.projectTools,
		AgentBehaviorMainInstruction: AgentBehaviorMainInstruction,
		AgentBehaviorInstructions:    buildAgentBehaviorInstructions(sections),
	})
}

var agentBehaviorInstructionDomains = []string{
	"structure",
	"git",
	"parent",
	"deps",
	"container",
	"tools",
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
	repoStructure     json.RawMessage
	gitOverview       json.RawMessage
	workspaceContext  json.RawMessage
	dependencies      json.RawMessage
	containerOverview json.RawMessage
	projectTools      json.RawMessage
}

func shouldIncludeBehaviorHint(domainName string, sections exploreSections) bool {
	switch domainName {
	case "structure":
		var structure struct {
			EntryCount int `json:"entryCount"`
		}
		return json.Unmarshal(sections.repoStructure, &structure) == nil && structure.EntryCount > 0
	case "git":
		var git struct {
			IsGitRepo bool `json:"isGitRepo"`
		}
		return json.Unmarshal(sections.gitOverview, &git) == nil && git.IsGitRepo
	case "parent":
		var parent struct {
			GitSiblingProjects []string `json:"gitSiblingProjects"`
			SiblingProjects    []string `json:"siblingProjects"`
		}
		if json.Unmarshal(sections.workspaceContext, &parent) != nil {
			return false
		}
		return len(parent.GitSiblingProjects) > 0 || len(parent.SiblingProjects) > 0
	case "deps":
		var deps []string
		return json.Unmarshal(sections.dependencies, &deps) == nil && len(deps) > 0
	case "container":
		return hasContainerOverviewData(sections.containerOverview)
	case "tools":
		return hasProjectToolsData(sections.projectTools)
	default:
		return false
	}
}

func hasContainerOverviewData(containerOverview json.RawMessage) bool {
	if len(containerOverview) == 0 {
		return false
	}

	var overview struct {
		DetectedContainerFileCount int `json:"detectedContainerFileCount"`
		RunningContainerCount      int `json:"runningContainerCount"`
		AvailableContainerCLICount int `json:"availableContainerCLICount"`
	}
	if json.Unmarshal(containerOverview, &overview) != nil {
		return false
	}

	return overview.DetectedContainerFileCount > 0 ||
		overview.RunningContainerCount > 0 ||
		overview.AvailableContainerCLICount > 0
}

func hasProjectToolsData(projectTools json.RawMessage) bool {
	if len(projectTools) == 0 {
		return false
	}

	var tools struct {
		MakefileTargetCount    int `json:"makefileTargetCount"`
		PackageJsonScriptCount int `json:"packageJsonScriptCount"`
		ShellScriptCount       int `json:"shellScriptCount"`
	}
	if json.Unmarshal(projectTools, &tools) != nil {
		return false
	}

	return tools.MakefileTargetCount > 0 || tools.PackageJsonScriptCount > 0 || tools.ShellScriptCount > 0
}

func overviewSection(fn func() resource.OverviewResource, verbose bool) (json.RawMessage, error) {
	result, err := fn()(verbose)
	if err != nil {
		return nil, err
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
