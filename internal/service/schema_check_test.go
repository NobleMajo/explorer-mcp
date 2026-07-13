package service

import (
	"encoding/json"
	"strings"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestExploreToolInputSchema(t *testing.T) {
	t.Parallel()

	if exploreToolInputSchema.Properties == nil || exploreToolInputSchema.Properties["projectRootPath"] == nil {
		t.Fatalf("schema missing projectRootPath property: %+v", exploreToolInputSchema)
	}

	foundRequired := false
	for _, name := range exploreToolInputSchema.Required {
		if name == "projectRootPath" {
			foundRequired = true
			break
		}
	}
	if !foundRequired {
		t.Fatalf("projectRootPath not required: %v", exploreToolInputSchema.Required)
	}
}

func TestRegisterExploreToolSchema(t *testing.T) {
	t.Parallel()

	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "test", Version: "1"}, nil)
	registerExploreTool(server, testExploreSettingsAllSections(false))

	// Re-register via same path used at runtime and inspect internal tool list through client.
	ctx := t.Context()
	clientTransport, serverTransport := mcpsdk.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test-client", Version: "1"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	t.Cleanup(func() { _ = clientSession.Close() })

	result, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	var exploreTool *mcpsdk.Tool
	for _, tool := range result.Tools {
		if tool.Name == "explore" {
			exploreTool = tool
			break
		}
	}
	if exploreTool == nil {
		t.Fatal("explore tool not found")
	}

	raw, err := json.Marshal(exploreTool.InputSchema)
	if err != nil {
		t.Fatalf("marshal input schema: %v", err)
	}

	var decoded struct {
		Type       string                     `json:"type"`
		Properties map[string]json.RawMessage `json:"properties"`
		Required   []string                   `json:"required"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal input schema: %v", err)
	}
	if decoded.Properties["projectRootPath"] == nil {
		t.Fatalf("projectRootPath missing from tool schema: %s", raw)
	}

	foundRequired := false
	for _, name := range decoded.Required {
		if name == "projectRootPath" {
			foundRequired = true
			break
		}
	}
	if !foundRequired {
		t.Fatalf("projectRootPath not required in tool schema: %s", raw)
	}
}

func TestExploreToolRequiresProjectRootPath(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "test", Version: "1"}, nil)
	registerExploreTool(server, testExploreSettingsAllSections(false))

	clientTransport, serverTransport := mcpsdk.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test-client", Version: "1"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	t.Cleanup(func() { _ = clientSession.Close() })

	result, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
		Name:      "explore",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected validation error for missing projectRootPath")
	}
	if len(result.Content) == 0 || !containsText(result.Content[0], "projectRootPath") {
		t.Fatalf("expected projectRootPath validation error, got %+v", result.Content)
	}
}

func containsText(content mcpsdk.Content, needle string) bool {
	text, ok := content.(*mcpsdk.TextContent)
	if !ok {
		return false
	}
	return strings.Contains(text.Text, needle)
}
