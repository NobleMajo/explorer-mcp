package gh

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestFormatOrgLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{name: "multiline", raw: "acme\n\nwidgets\n", want: []string{"acme", "widgets"}},
		{name: "empty", raw: "", want: []string{}},
		{name: "whitespace only", raw: "  \n\t\n", want: []string{}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := formatOrgLines(tc.raw)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("formatOrgLines() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestRepoNameFromRemoteURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		raw  string
		want string
	}{
		{raw: "https://github.com/NobleMajo/explorer-mcp.git", want: "explorer-mcp"},
		{raw: "git@github.com:NobleMajo/explorer-mcp.git", want: "explorer-mcp"},
		{raw: "ssh://git@github.com/NobleMajo/explorer-mcp.git", want: "explorer-mcp"},
		{raw: "https://github.com/NobleMajo/explorer-mcp", want: "explorer-mcp"},
		{raw: "", want: ""},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.raw, func(t *testing.T) {
			t.Parallel()
			if got := repoNameFromRemoteURL(tc.raw); got != tc.want {
				t.Fatalf("repoNameFromRemoteURL(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestOwnerAndNameFromURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		raw       string
		wantOwner string
		wantName  string
	}{
		{raw: "https://github.com/acme/demo.git", wantOwner: "acme", wantName: "demo"},
		{raw: "git@github.com:acme/demo.git", wantOwner: "acme", wantName: "demo"},
		{raw: "ssh://git@github.com/acme/demo.git", wantOwner: "acme", wantName: "demo"},
		{raw: "demo", wantOwner: "", wantName: "demo"},
		{raw: "", wantOwner: "", wantName: ""},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.raw, func(t *testing.T) {
			t.Parallel()
			owner, name := ownerAndNameFromURL(tc.raw)
			if owner != tc.wantOwner || name != tc.wantName {
				t.Fatalf("ownerAndNameFromURL(%q) = (%q,%q), want (%q,%q)", tc.raw, owner, name, tc.wantOwner, tc.wantName)
			}
		})
	}
}

func TestFilterMatchingRepos(t *testing.T) {
	t.Parallel()

	repos := []ghRepo{
		{Name: "explorer-mcp", ID: "1", URL: "https://github.com/acme/explorer-mcp", Description: "tool"},
		{Name: "other", ID: "2", URL: "https://github.com/acme/other", Description: "nope"},
		{Name: "demo", ID: "3", URL: "https://github.com/acme/demo", Description: "contains Explorer-MCP here"},
		{Name: "by-id", ID: "explorer-mcp-id", URL: "https://github.com/acme/by-id", Description: ""},
		{Name: "by-url", ID: "4", URL: "https://github.com/acme/has-explorer-mcp-path", Description: ""},
	}

	got := filterMatchingRepos(repos, []string{"explorer-mcp"})
	if len(got) != 4 {
		t.Fatalf("len = %d, want 4 (%#v)", len(got), got)
	}

	if filterMatchingRepos(repos, nil) != nil {
		t.Fatal("expected nil matches for empty needles")
	}
}

func TestFormatRepoLine(t *testing.T) {
	t.Parallel()

	got := formatRepoLine(ghRepo{
		Name: "explorer-mcp",
		ID:   "R_123",
		URL:  "https://github.com/NobleMajo/explorer-mcp",
	})
	want := "NobleMajo/explorer-mcp @R_123 @https://github.com/NobleMajo/explorer-mcp"
	if got != want {
		t.Fatalf("formatRepoLine() = %q, want %q", got, want)
	}

	fallback := formatRepoLine(ghRepo{Name: "solo", ID: 42.0, URL: ""})
	if fallback != "unknown/solo @42 @" {
		t.Fatalf("formatRepoLine fallback = %q", fallback)
	}
}

func TestRepoIDString(t *testing.T) {
	t.Parallel()

	if got := repoIDString(nil); got != "" {
		t.Fatalf("nil = %q", got)
	}
	if got := repoIDString("abc"); got != "abc" {
		t.Fatalf("string = %q", got)
	}
	if got := repoIDString(float64(99)); got != "99" {
		t.Fatalf("float = %q", got)
	}
}

func TestParseRepoListJSON(t *testing.T) {
	t.Parallel()

	got := parseRepoListJSON(`[{"name":"demo","id":"1","url":"https://github.com/acme/demo","description":"x"}]`)
	if len(got) != 1 || got[0].Name != "demo" {
		t.Fatalf("parseRepoListJSON() = %#v", got)
	}
	if parseRepoListJSON("{bad") != nil {
		t.Fatal("expected nil for invalid json")
	}
	if parseRepoListJSON("") != nil {
		t.Fatal("expected nil for empty")
	}
	numeric := parseRepoListJSON(`[{"name":"n","id":7,"url":"https://github.com/a/n","description":""}]`)
	if len(numeric) != 1 || repoIDString(numeric[0].ID) != "7" {
		t.Fatalf("numeric id parse = %#v", numeric)
	}
}

func TestRawCacheGetOrFetch(t *testing.T) {
	t.Parallel()

	cache := newRawCache()
	calls := 0
	fetch := func() (string, error) {
		calls++
		return "raw", nil
	}

	raw, err := cache.getOrFetch("k", time.Minute, fetch)
	if err != nil || raw != "raw" || calls != 1 {
		t.Fatalf("first fetch raw=%q err=%v calls=%d", raw, err, calls)
	}

	raw, err = cache.getOrFetch("k", time.Minute, fetch)
	if err != nil || raw != "raw" || calls != 1 {
		t.Fatalf("cached fetch raw=%q err=%v calls=%d", raw, err, calls)
	}
}

func TestRawCacheExpires(t *testing.T) {
	t.Parallel()

	cache := newRawCache()
	calls := 0
	fetch := func() (string, error) {
		calls++
		return "v", nil
	}

	if _, err := cache.getOrFetch("k", time.Millisecond, fetch); err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Millisecond)
	if _, err := cache.getOrFetch("k", time.Millisecond, fetch); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2 after expiry", calls)
	}
}

func TestRawCacheFetchError(t *testing.T) {
	t.Parallel()

	cache := newRawCache()
	_, err := cache.getOrFetch("k", time.Minute, func() (string, error) {
		return "", errors.New("boom")
	})
	if err == nil {
		t.Fatal("expected fetch error")
	}
}

func withGhTestHooks(t *testing.T) {
	t.Helper()
	oldLookPath := lookPath
	oldRunGh := runGh
	oldRunGit := runGit
	oldOrgCache := orgListCache
	oldRepoCache := repoListCache
	t.Cleanup(func() {
		lookPath = oldLookPath
		runGh = oldRunGh
		runGit = oldRunGit
		orgListCache = oldOrgCache
		repoListCache = oldRepoCache
	})
	orgListCache = newRawCache()
	repoListCache = newRawCache()
}

func TestGhOverviewOmitsWhenGhMissing(t *testing.T) {
	withGhTestHooks(t)
	lookPath = func(file string) (string, error) {
		return "", errors.New("missing")
	}

	result, err := GhOverview()(t.TempDir(), false)
	if err != nil {
		t.Fatalf("GhOverview() error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil when gh missing, got %#v", result)
	}
}

func TestGhOverviewOmitsWhenNotLoggedIn(t *testing.T) {
	withGhTestHooks(t)
	lookPath = func(file string) (string, error) { return "/bin/gh", nil }
	runGh = func(dir string, args ...string) (string, error) {
		return "", errors.New("not logged in")
	}

	result, err := GhOverview()(t.TempDir(), false)
	if err != nil {
		t.Fatalf("GhOverview() error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil when not logged in, got %#v", result)
	}
}

func TestGhOverviewMatchingRepos(t *testing.T) {
	withGhTestHooks(t)

	project := filepath.Join(t.TempDir(), "explorer-mcp")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}

	var sawUserLimit, sawOrgLimit bool
	lookPath = func(file string) (string, error) { return "/bin/gh", nil }
	runGit = func(dir string, args ...string) (string, error) {
		return "origin\tgit@github.com:NobleMajo/explorer-mcp.git (fetch)\norigin\tgit@github.com:NobleMajo/explorer-mcp.git (push)\n", nil
	}
	runGh = func(dir string, args ...string) (string, error) {
		switch {
		case len(args) >= 2 && args[0] == "auth" && args[1] == "status":
			return "", nil
		case len(args) >= 2 && args[0] == "org" && args[1] == "list":
			return "acme\n", nil
		case len(args) >= 4 && args[0] == "repo" && args[1] == "list" && args[2] == "--limit":
			if args[3] != "1000" {
				t.Fatalf("user repo list limit = %q, want 1000", args[3])
			}
			sawUserLimit = true
			return `[{"name":"explorer-mcp","id":"1","url":"https://github.com/me/explorer-mcp","description":"user"},{"name":"dup","id":"1","url":"https://github.com/me/explorer-mcp","description":"dup id"}]`, nil
		case len(args) >= 5 && args[0] == "repo" && args[1] == "list" && args[2] == "acme" && args[3] == "--limit":
			if args[4] != "1000" {
				t.Fatalf("org repo list limit = %q, want 1000", args[4])
			}
			sawOrgLimit = true
			return `[{"name":"other","id":"2","url":"https://github.com/acme/other","description":"nope"},{"name":"widgets","id":"3","url":"https://github.com/acme/widgets","description":"mentions explorer-mcp"}]`, nil
		default:
			return "", errors.New("unexpected gh args: " + strings.Join(args, " "))
		}
	}

	result, err := GhOverview()(project, false)
	if err != nil {
		t.Fatalf("GhOverview() error: %v", err)
	}
	if !sawUserLimit || !sawOrgLimit {
		t.Fatalf("expected --limit 1000 on user and org repo list (user=%v org=%v)", sawUserLimit, sawOrgLimit)
	}

	lines, ok := result.([]string)
	if !ok || len(lines) != 2 {
		t.Fatalf("result = %#v", result)
	}
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "me/explorer-mcp @1 @") {
		t.Fatalf("missing user match in %q", joined)
	}
	if !strings.Contains(joined, "acme/widgets @3 @") {
		t.Fatalf("missing org description match in %q", joined)
	}
}

func TestGhOverviewCachesOrgAndRepoLists(t *testing.T) {
	withGhTestHooks(t)

	project := filepath.Join(t.TempDir(), "demo")
	testutil.WriteFile(t, filepath.Join(project, ".keep"), "")

	calls := map[string]int{}
	lookPath = func(file string) (string, error) { return "/bin/gh", nil }
	runGit = func(dir string, args ...string) (string, error) { return "", nil }
	runGh = func(dir string, args ...string) (string, error) {
		key := strings.Join(args, " ")
		calls[key]++
		switch {
		case strings.HasPrefix(key, "auth status"):
			return "", nil
		case key == "org list":
			return "acme", nil
		case strings.HasPrefix(key, "repo list --limit"):
			return `[{"name":"demo","id":"1","url":"https://github.com/me/demo","description":""}]`, nil
		case strings.HasPrefix(key, "repo list acme --limit"):
			return `[]`, nil
		default:
			return "", errors.New("unexpected " + key)
		}
	}

	if _, err := GhOverview()(project, false); err != nil {
		t.Fatalf("first: %v", err)
	}
	if _, err := GhOverview()(project, false); err != nil {
		t.Fatalf("second: %v", err)
	}

	if calls["org list"] != 1 {
		t.Fatalf("org list calls = %d, want 1", calls["org list"])
	}
	userCalls := 0
	orgCalls := 0
	for key, n := range calls {
		if strings.HasPrefix(key, "repo list --limit") {
			userCalls += n
		}
		if strings.HasPrefix(key, "repo list acme --limit") {
			orgCalls += n
		}
	}
	if userCalls != 1 || orgCalls != 1 {
		t.Fatalf("repo list calls user=%d org=%d want 1 each; all=%v", userCalls, orgCalls, calls)
	}
}

func TestGhOverviewOrgListErrorStillListsUserRepos(t *testing.T) {
	withGhTestHooks(t)

	project := filepath.Join(t.TempDir(), "demo")
	testutil.WriteFile(t, filepath.Join(project, ".keep"), "")

	lookPath = func(file string) (string, error) { return "/bin/gh", nil }
	runGit = func(dir string, args ...string) (string, error) { return "", nil }
	runGh = func(dir string, args ...string) (string, error) {
		switch {
		case len(args) >= 2 && args[0] == "auth" && args[1] == "status":
			return "", nil
		case len(args) >= 2 && args[0] == "org" && args[1] == "list":
			return "", errors.New("org list failed")
		case len(args) >= 2 && args[0] == "repo" && args[1] == "list":
			return `[{"name":"demo","id":"1","url":"https://github.com/me/demo","description":""}]`, nil
		default:
			return "", errors.New("unexpected")
		}
	}

	result, err := GhOverview()(project, false)
	if err != nil {
		t.Fatalf("GhOverview() error: %v", err)
	}
	lines, ok := result.([]string)
	if !ok || len(lines) != 1 || !strings.Contains(lines[0], "me/demo @1 @") {
		t.Fatalf("result = %#v", result)
	}
}

func TestGhOverviewNoMatchMessage(t *testing.T) {
	withGhTestHooks(t)

	project := filepath.Join(t.TempDir(), "explorer-mcp")
	testutil.WriteFile(t, filepath.Join(project, ".keep"), "")

	lookPath = func(file string) (string, error) { return "/bin/gh", nil }
	runGit = func(dir string, args ...string) (string, error) {
		return "", errors.New("no remotes")
	}
	runGh = func(dir string, args ...string) (string, error) {
		switch {
		case len(args) >= 2 && args[0] == "auth" && args[1] == "status":
			return "", nil
		case len(args) >= 2 && args[0] == "org" && args[1] == "list":
			return "", nil
		case len(args) >= 2 && args[0] == "repo" && args[1] == "list":
			return `[{"name":"unrelated","id":"9","url":"https://github.com/me/unrelated","description":"x"}]`, nil
		default:
			return "", errors.New("unexpected")
		}
	}

	result, err := GhOverview()(project, false)
	if err != nil {
		t.Fatalf("GhOverview() error: %v", err)
	}
	lines, ok := result.([]string)
	if !ok || len(lines) != 1 || lines[0] != noSimilarReposMessage {
		t.Fatalf("result = %#v", result)
	}
}

func TestCollectNeedlesUsesDirAndRemotes(t *testing.T) {
	withGhTestHooks(t)

	project := filepath.Join(t.TempDir(), "LocalName")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}

	runGit = func(dir string, args ...string) (string, error) {
		return "origin\thttps://github.com/acme/remote-name.git (fetch)\nupstream\tgit@github.com:acme/LocalName.git (fetch)\n", nil
	}

	needles := collectNeedles(project)
	joined := strings.ToLower(strings.Join(needles, ","))
	if !strings.Contains(joined, "localname") || !strings.Contains(joined, "remote-name") {
		t.Fatalf("needles = %#v", needles)
	}
	// Dedup case-insensitive LocalName from dir and remote.
	count := 0
	for _, n := range needles {
		if strings.EqualFold(n, "LocalName") {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("LocalName appearances = %d, want 1 in %#v", count, needles)
	}
}
