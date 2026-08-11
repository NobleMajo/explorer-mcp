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

	// bash * ask => collapses to just "ask"
	if resp.Permissions["bash"] != "ask" {
		t.Fatalf("permissions[bash] = %v, want \"ask\" (collapsed single wildcard)", resp.Permissions["bash"])
	}

	// read *.env ask => single non-trivial pattern under one action => map with string pattern
	readPerms, ok := resp.Permissions["read"].(map[string]any)
	if !ok {
		t.Fatalf("permissions[read] type = %T, want map[string]any", resp.Permissions["read"])
	}
	if readPerms["ask"] != "*.env" {
		t.Fatalf("permissions[read][ask] = %v, want \"*.env\"", readPerms["ask"])
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
	// * '*'=allow => collapsed to string "allow"
	if len(resp.Permissions) != 1 {
		t.Fatalf("permissions len = %d, want 1", len(resp.Permissions))
	}
	if resp.Permissions["*"] != "allow" {
		t.Fatalf("permissions[*] = %v, want \"allow\" (collapsed trivial)", resp.Permissions["*"])
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

// --- compaction unit tests ---

func TestCompactPermissionsNilForEmptyInput(t *testing.T) {
	if compactPermissions(nil) != nil {
		t.Fatal("expected nil for nil input")
	}
	if compactPermissions([]permissionRule{}) != nil {
		t.Fatal("expected nil for empty slice")
	}
}

func TestDeduplicateRulesLastWins(t *testing.T) {
	rules := []permissionRule{
		{Permission: "question", Pattern: "*", Action: "deny"},
		{Permission: "read", Pattern: "*.env", Action: "ask"},
		{Permission: "question", Pattern: "*", Action: "allow"},
	}
	got := deduplicateRules(rules)
	if len(got) != 2 {
		t.Fatalf("deduplicateRules len = %d, want 2; got %v", len(got), got)
	}
	for _, r := range got {
		if r.Permission == "question" && r.Action != "allow" {
			t.Fatalf("question rule action = %q, want allow (last-wins)", r.Action)
		}
	}
}

func TestDeduplicateRulesPreservesFirstSeenOrder(t *testing.T) {
	rules := []permissionRule{
		{Permission: "bash", Pattern: "git *", Action: "allow"},
		{Permission: "read", Pattern: "*", Action: "allow"},
		{Permission: "bash", Pattern: "git *", Action: "ask"},
	}
	got := deduplicateRules(rules)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Permission != "bash" {
		t.Fatalf("first rule should be bash, got %q", got[0].Permission)
	}
	if got[0].Action != "ask" {
		t.Fatalf("bash action = %q, want ask (last-wins)", got[0].Action)
	}
}

func TestCompactPatternsSingleReturnsString(t *testing.T) {
	got := compactPatterns([]string{"*.env"})
	if got != "*.env" {
		t.Fatalf("compactPatterns = %v, want \"*.env\"", got)
	}
}

func TestCompactPatternsMultipleReturnsSortedSlice(t *testing.T) {
	got := compactPatterns([]string{"git log *", "cat *", "docker ps *"})
	list, ok := got.([]string)
	if !ok {
		t.Fatalf("type = %T, want []string", got)
	}
	want := []string{"cat *", "docker ps *", "git log *"}
	if !slices.Equal(list, want) {
		t.Fatalf("compactPatterns = %v, want %v", list, want)
	}
}

func TestCompactPermissionCollapsesBareStar(t *testing.T) {
	got := compactPermission(map[string][]string{"ask": {"*"}})
	if got != "ask" {
		t.Fatalf("compactPermission = %v, want \"ask\"", got)
	}
}

func TestCompactPermissionKeepsMultiAction(t *testing.T) {
	got := compactPermission(map[string][]string{
		"allow": {"*"},
		"deny":  {"*.env"},
	})
	m, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("type = %T, want map", got)
	}
	if m["allow"] != "*" {
		t.Fatalf("allow = %v, want \"*\"", m["allow"])
	}
	if m["deny"] != "*.env" {
		t.Fatalf("deny = %v, want \"*.env\"", m["deny"])
	}
}

func TestCompactPermissionKeepsNonTrivialSingleAction(t *testing.T) {
	got := compactPermission(map[string][]string{"allow": {"*.ts", "*.go"}})
	m, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("type = %T, want map (non-trivial should not collapse)", got)
	}
	list, ok := m["allow"].([]string)
	if !ok {
		t.Fatalf("allow type = %T, want []string", m["allow"])
	}
	if !slices.Equal(list, []string{"*.go", "*.ts"}) {
		t.Fatalf("allow = %v, want [*.go *.ts]", list)
	}
}

func TestCompactPermissionsGroupsAllTypes(t *testing.T) {
	rules := []permissionRule{
		{Permission: "bash", Pattern: "git log *", Action: "allow"},
		{Permission: "bash", Pattern: "git diff *", Action: "allow"},
		{Permission: "bash", Pattern: "rm -rf *", Action: "deny"},
		{Permission: "bash", Pattern: "*", Action: "ask"},
		{Permission: "read", Pattern: "*.env", Action: "ask"},
		{Permission: "skill", Pattern: "use-explorer-mcp", Action: "allow"},
		{Permission: "skill", Pattern: "caveman", Action: "allow"},
		{Permission: "*", Pattern: "*", Action: "ask"},
	}
	result := compactPermissions(rules)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	for _, perm := range []string{"bash", "read", "skill", "*"} {
		if result[perm] == nil {
			t.Errorf("missing permission type %q", perm)
		}
	}

	// bash keeps globs grouped by action — no regex
	bashPerms, ok := result["bash"].(map[string]any)
	if !ok {
		t.Fatalf("bash type = %T, want map[string]any", result["bash"])
	}
	bashAllow, ok := bashPerms["allow"].([]string)
	if !ok {
		t.Fatalf("bash.allow type = %T, want []string", bashPerms["allow"])
	}
	if !slices.Equal(bashAllow, []string{"git diff *", "git log *"}) {
		t.Fatalf("bash.allow = %v, want [git diff * git log *]", bashAllow)
	}
	if bashPerms["deny"] != "rm -rf *" {
		t.Fatalf("bash.deny = %v, want \"rm -rf *\"", bashPerms["deny"])
	}
	if bashPerms["ask"] != "*" {
		t.Fatalf("bash.ask = %v, want \"*\"", bashPerms["ask"])
	}

	skillPerms, ok := result["skill"].(map[string]any)
	if !ok {
		t.Fatalf("skill type = %T, want map[string]any", result["skill"])
	}
	skillAllow, ok := skillPerms["allow"].([]string)
	if !ok {
		t.Fatalf("skill.allow type = %T, want []string", skillPerms["allow"])
	}
	if !slices.Equal(skillAllow, []string{"caveman", "use-explorer-mcp"}) {
		t.Fatalf("skill.allow = %v, want [caveman use-explorer-mcp]", skillAllow)
	}

	// * '*'=ask => collapsed to string "ask"
	if result["*"] != "ask" {
		t.Fatalf("permissions[*] = %v, want \"ask\" (collapsed trivial)", result["*"])
	}
}

func TestCompactPermissionsDeduplicatesBeforeCompacting(t *testing.T) {
	rules := []permissionRule{
		{Permission: "question", Pattern: "*", Action: "deny"},
		{Permission: "question", Pattern: "*", Action: "allow"},
	}
	result := compactPermissions(rules)

	// After last-wins dedup: only allow survives with pattern * => collapses to "allow"
	if result["question"] != "allow" {
		t.Fatalf("question = %v, want \"allow\" (last-wins dedup + collapse)", result["question"])
	}
}

func TestCompactPermissionsBashKeepsOverlappingGlobsSeparate(t *testing.T) {
	// bun * deny and bun run * allow must both survive as readable globs —
	// do not merge unrelated patterns into one blob.
	rules := []permissionRule{
		{Permission: "bash", Pattern: "*", Action: "ask"},
		{Permission: "bash", Pattern: "bun *", Action: "deny"},
		{Permission: "bash", Pattern: "bun run *", Action: "allow"},
		{Permission: "bash", Pattern: "cargo *", Action: "deny"},
		{Permission: "bash", Pattern: "*\n*", Action: "deny"},
		{Permission: "bash", Pattern: "cat *", Action: "allow"},
	}
	result := compactPermissions(rules)

	bashPerms, ok := result["bash"].(map[string]any)
	if !ok {
		t.Fatalf("bash type = %T, want map[string]any", result["bash"])
	}

	deny, ok := bashPerms["deny"].([]string)
	if !ok {
		t.Fatalf("bash.deny type = %T, want []string", bashPerms["deny"])
	}
	if !slices.Equal(deny, []string{"*\n*", "bun *", "cargo *"}) {
		t.Fatalf("bash.deny = %v, want [*$\\n* bun * cargo *]", deny)
	}

	allow, ok := bashPerms["allow"].([]string)
	if !ok {
		t.Fatalf("bash.allow type = %T, want []string", bashPerms["allow"])
	}
	if !slices.Equal(allow, []string{"bun run *", "cat *"}) {
		t.Fatalf("bash.allow = %v, want [bun run * cat *]", allow)
	}

	if bashPerms["ask"] != "*" {
		t.Fatalf("bash.ask = %v, want \"*\"", bashPerms["ask"])
	}
}

func TestCompactPermissionsLastWinsAcrossDuplicatePatterns(t *testing.T) {
	rules := []permissionRule{
		{Permission: "bash", Pattern: "make *", Action: "ask"},
		{Permission: "bash", Pattern: "npm run *", Action: "ask"},
		{Permission: "bash", Pattern: "make *", Action: "allow"},
		{Permission: "bash", Pattern: "npm run *", Action: "allow"},
	}
	result := compactPermissions(rules)

	bashPerms, ok := result["bash"].(map[string]any)
	if !ok {
		t.Fatalf("bash type = %T, want map[string]any", result["bash"])
	}
	if _, hasAsk := bashPerms["ask"]; hasAsk {
		t.Fatalf("ask should be gone after last-wins, got %v", bashPerms)
	}
	allow, ok := bashPerms["allow"].([]string)
	if !ok {
		t.Fatalf("bash.allow type = %T, want []string", bashPerms["allow"])
	}
	if !slices.Equal(allow, []string{"make *", "npm run *"}) {
		t.Fatalf("bash.allow = %v, want [make * npm run *]", allow)
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
