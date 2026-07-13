package opencode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

const (
	cliName   = "opencode"
	agentName = "build"
)

var nativeToolIDs = map[string]struct{}{
	"invalid":     {},
	"question":    {},
	"bash":        {},
	"read":        {},
	"glob":        {},
	"grep":        {},
	"edit":        {},
	"write":       {},
	"task":        {},
	"webfetch":    {},
	"todowrite":   {},
	"websearch":   {},
	"skill":       {},
	"lsp":         {},
	"plan_enter":  {},
	"plan_exit":   {},
	"execute":     {},
}

type permissionRule struct {
	Permission string `json:"permission"`
	Action     string `json:"action"`
	Pattern    string `json:"pattern"`
}

type debugAgentResponse struct {
	Permission []permissionRule  `json:"permission"`
	Tools      map[string]bool   `json:"tools"`
}

type opencodeOverviewResponse struct {
	Permissions []string `json:"permissions,omitempty"`
	MCP         []string `json:"mcp,omitempty"`
}

func buildOpencodeOverview(verbose bool) (any, error) {
	_ = verbose

	if _, err := exec.LookPath(cliName); err != nil {
		return nil, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(cliName, "debug", "agent", agentName)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, nil
	}

	out = bytes.TrimSpace(out)
	if len(out) == 0 {
		return nil, nil
	}

	var parsed debugAgentResponse
	if err := json.Unmarshal(out, &parsed); err != nil {
		return nil, nil
	}

	return opencodeOverviewResponse{
		Permissions: formatPermissions(parsed.Permission),
		MCP:         extractMCPServers(parsed.Tools),
	}, nil
}

func formatPermissions(rules []permissionRule) []string {
	if len(rules) == 0 {
		return nil
	}

	formatted := make([]string, 0, len(rules))
	for _, rule := range rules {
		formatted = append(formatted, fmt.Sprintf("%s '%s'=%s", rule.Permission, rule.Pattern, rule.Action))
	}
	return formatted
}

func extractMCPServers(tools map[string]bool) []string {
	if len(tools) == 0 {
		return nil
	}

	nonNative := make([]string, 0)
	for id, enabled := range tools {
		if !enabled {
			continue
		}
		if _, native := nativeToolIDs[id]; native {
			continue
		}
		if !strings.Contains(id, "_") {
			continue
		}
		nonNative = append(nonNative, id)
	}
	if len(nonNative) == 0 {
		return nil
	}

	sort.Strings(nonNative)

	servers := make(map[string]struct{})
	for _, id := range nonNative {
		parts := strings.Split(id, "_")
		for i := len(parts) - 1; i >= 1; i-- {
			prefix := strings.Join(parts[:i], "_")
			if matchesToolPrefix(nonNative, prefix) {
				servers[prefix] = struct{}{}
				break
			}
		}
	}

	if len(servers) == 0 {
		return nil
	}

	names := make([]string, 0, len(servers))
	for name := range servers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func matchesToolPrefix(toolIDs []string, prefix string) bool {
	for _, id := range toolIDs {
		if id == prefix || strings.HasPrefix(id, prefix+"_") {
			return true
		}
	}
	return false
}
