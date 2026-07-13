package git

import (
	"os/exec"
	"reflect"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestParseGitStatusShort(t *testing.T) {
	t.Parallel()

	t.Run("populated", func(t *testing.T) {
		t.Parallel()
		input := " M main.go\n?? new.txt\n"
		got := parseGitStatusShort(input)
		want := []gitChangedFile{
			{StatusCode: " M", FilePath: "main.go"},
			{StatusCode: "??", FilePath: "new.txt"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("parseGitStatusShort() = %#v, want %#v", got, want)
		}
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		if len(parseGitStatusShort("")) != 0 {
			t.Fatal("expected empty slice")
		}
	})

	t.Run("short line", func(t *testing.T) {
		t.Parallel()
		got := parseGitStatusShort("?\n")
		if len(got) != 1 || got[0].FilePath != "" {
			t.Fatalf("parseGitStatusShort(?) = %#v", got)
		}
	})
}

func TestParseGitLog(t *testing.T) {
	t.Parallel()

	t.Run("populated", func(t *testing.T) {
		t.Parallel()
		input := "abc1234 first commit\ndef5678 second commit\n"
		got := parseGitLog(input)
		want := []gitCommit{
			{ShortCommitHash: "abc1234", CommitSubject: "first commit"},
			{ShortCommitHash: "def5678", CommitSubject: "second commit"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("parseGitLog() = %#v, want %#v", got, want)
		}
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		if len(parseGitLog("")) != 0 {
			t.Fatal("expected empty slice")
		}
	})
}

func TestGitOverviewWithoutGitRepo(t *testing.T) {
	root := t.TempDir()
	testutil.Chdir(t, root)

	result, err := GitOverview()(false)
	if err != nil {
		t.Fatalf("GitOverview() error: %v", err)
	}

	resp, ok := result.(gitOverviewResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	if resp.ToolName != "git_overview" {
		t.Fatalf("toolName = %q", resp.ToolName)
	}
	if resp.IsGitRepo {
		t.Fatal("expected isGitRepo false outside git repo")
	}
	if len(resp.ChangedFiles) != 0 || len(resp.RecentCommits) != 0 {
		t.Fatal("expected empty changed files and commits")
	}
}

func TestGitOverviewGitNotInPath(t *testing.T) {
	root := t.TempDir()
	testutil.Chdir(t, root)
	t.Setenv("PATH", t.TempDir())

	result, err := GitOverview()(false)
	if err != nil {
		t.Fatalf("GitOverview() error: %v", err)
	}

	resp, ok := result.(gitOverviewResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if resp.IsGitAvailable || resp.ErrorMessage == "" {
		t.Fatalf("expected git unavailable response, got %+v", resp)
	}
}

func TestGitOverviewInsideGitRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not in PATH")
	}

	root := t.TempDir()
	testutil.Chdir(t, root)

	runGit(t, root, "init")
	testutil.WriteFile(t, root+"/README.md", "demo\n")

	result, err := GitOverview()(false)
	if err != nil {
		t.Fatalf("GitOverview() error: %v", err)
	}

	resp, ok := result.(gitOverviewResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	if !resp.IsGitAvailable || !resp.IsGitRepo || !resp.IsInsideWorkTree {
		t.Fatalf("unexpected git flags: %+v", resp)
	}
}

func TestGitOverviewReportsRecentCommits(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not in PATH")
	}

	root := t.TempDir()
	testutil.Chdir(t, root)

	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "test")
	testutil.WriteFile(t, root+"/README.md", "demo\n")
	runGit(t, root, "add", "README.md")
	runGit(t, root, "commit", "-m", "init")

	result, err := GitOverview()(false)
	if err != nil {
		t.Fatalf("GitOverview() error: %v", err)
	}

	resp, ok := result.(gitOverviewResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if resp.RecentCommitCount == 0 || len(resp.RecentCommits) == 0 {
		t.Fatalf("expected recent commits, got %+v", resp)
	}
}

func TestGitOverviewDirtyWorkingTree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not in PATH")
	}

	root := t.TempDir()
	testutil.Chdir(t, root)

	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "test")
	testutil.WriteFile(t, root+"/README.md", "demo\n")
	runGit(t, root, "add", "README.md")
	runGit(t, root, "commit", "-m", "init")
	testutil.WriteFile(t, root+"/README.md", "changed\n")

	result, err := GitOverview()(false)
	if err != nil {
		t.Fatalf("GitOverview() error: %v", err)
	}

	resp, ok := result.(gitOverviewResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	if resp.IsWorkingTreeClean || resp.ChangedFileCount == 0 {
		t.Fatalf("expected dirty working tree, got %+v", resp)
	}
}

func TestGitOverviewDetachedHead(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not in PATH")
	}

	root := t.TempDir()
	testutil.Chdir(t, root)

	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "test")
	testutil.WriteFile(t, root+"/README.md", "demo\n")
	runGit(t, root, "add", "README.md")
	runGit(t, root, "commit", "-m", "init")
	runGit(t, root, "checkout", "--detach")

	result, err := GitOverview()(false)
	if err != nil {
		t.Fatalf("GitOverview() error: %v", err)
	}

	resp, ok := result.(gitOverviewResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	if !resp.IsDetachedHead || resp.DetachedHeadCommitHash == "" {
		t.Fatalf("expected detached HEAD, got %+v", resp)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}
