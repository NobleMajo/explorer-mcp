package opencode

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

const stubAgentJSON = `{"name":"build","permission":[{"permission":"bash","pattern":"*","action":"ask"},{"permission":"read","pattern":"*.env","action":"ask"}],"tools":{"bash":true,"read":true,"my_mcp_search":true,"my_mcp_list":true}}`

func writeOpencodeCLIStub(t *testing.T, binDir, body string) {
	t.Helper()
	path := filepath.Join(binDir, cliName)
	testutil.WriteFile(t, path, body)
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestOpencodeOverviewNilWhenCLINotInPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	result, err := OpencodeOverview()(t.TempDir(), false)
	if err != nil {
		t.Fatalf("OpencodeOverview() error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil when opencode not in PATH, got %+v", result)
	}
}

func TestOpencodeOverviewNilWhenDebugAgentFails(t *testing.T) {
	binDir := t.TempDir()
	writeOpencodeCLIStub(t, binDir, "#!/bin/sh\nexit 1\n")
	t.Setenv("PATH", binDir)

	result, err := OpencodeOverview()(t.TempDir(), false)
	if err != nil {
		t.Fatalf("OpencodeOverview() error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil when debug agent fails, got %+v", result)
	}
}

func TestOpencodeOverviewFromDebugAgentBuild(t *testing.T) {
	binDir := t.TempDir()
	writeOpencodeCLIStub(t, binDir, "#!/bin/sh\nif [ \"$1\" = debug ] && [ \"$2\" = agent ] && [ \"$3\" = build ]; then echo '"+stubAgentJSON+"'; fi\n")
	t.Setenv("PATH", binDir)

	result, err := OpencodeOverview()(t.TempDir(), false)
	if err != nil {
		t.Fatalf("OpencodeOverview() error: %v", err)
	}

	resp, ok := result.(opencodeOverviewResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	wantPermissions := []string{
		"bash '*'=ask",
		"read '*.env'=ask",
	}
	if !slices.Equal(resp.Permissions, wantPermissions) {
		t.Fatalf("permissions = %v, want %v", resp.Permissions, wantPermissions)
	}
	if !slices.Equal(resp.MCP, []string{"my_mcp"}) {
		t.Fatalf("mcp = %v, want [my_mcp]", resp.MCP)
	}
}

func TestOpencodeOverviewEmptyMCPWhenOnlyNativeTools(t *testing.T) {
	binDir := t.TempDir()
	writeOpencodeCLIStub(t, binDir, "#!/bin/sh\nif [ \"$1\" = debug ] && [ \"$2\" = agent ] && [ \"$3\" = build ]; then echo '{\"permission\":[{\"permission\":\"*\",\"pattern\":\"*\",\"action\":\"allow\"}],\"tools\":{\"bash\":true,\"read\":true}}'; fi\n")
	t.Setenv("PATH", binDir)

	result, err := OpencodeOverview()(t.TempDir(), false)
	if err != nil {
		t.Fatalf("OpencodeOverview() error: %v", err)
	}

	resp, ok := result.(opencodeOverviewResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if len(resp.Permissions) != 1 {
		t.Fatalf("permissions = %v, want one entry", resp.Permissions)
	}
	if resp.MCP != nil {
		t.Fatalf("mcp = %v, want nil", resp.MCP)
	}
}

func TestOpencodeOverviewUsesProjectCwdWhenRealCLIAvailable(t *testing.T) {
	if _, err := exec.LookPath(cliName); err != nil {
		t.Skip("opencode cli not installed")
	}

	root := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	testutil.Chdir(t, root)

	result, err := OpencodeOverview()(root, false)
	if err != nil {
		t.Fatalf("OpencodeOverview() error: %v", err)
	}
	if result == nil {
		t.Fatal("expected overview from real opencode debug agent build")
	}

	resp, ok := result.(opencodeOverviewResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if len(resp.Permissions) == 0 {
		t.Fatal("expected non-empty permissions from build agent defaults")
	}
}

func TestFormatPermissionsNilForEmptyInput(t *testing.T) {
	if formatPermissions(nil) != nil {
		t.Fatal("expected nil for empty permission rules")
	}
}

func TestExtractMCPServersGroupsToolPrefixes(t *testing.T) {
	servers := extractMCPServers(map[string]bool{
		"bash":          true,
		"my_mcp_search": true,
		"my_mcp_list":   true,
		"other_tool":    true,
	})
	if !slices.Equal(servers, []string{"my_mcp", "other"}) {
		t.Fatalf("servers = %v, want [my_mcp other]", servers)
	}
}
