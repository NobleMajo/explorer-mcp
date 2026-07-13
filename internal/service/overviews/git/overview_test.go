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
		want := []string{
			"m: main.go",
			"?: new.txt",
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("parseGitStatusShort() = %#v, want %#v", got, want)
		}
	})

	t.Run("staged and both modified", func(t *testing.T) {
		t.Parallel()
		input := "M  staged.go\nMM both.go\n D deleted.go\n"
		got := parseGitStatusShort(input)
		want := []string{
			"M: staged.go",
			"*: both.go",
			"d: deleted.go",
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
		if len(parseGitStatusShort("?\n")) != 0 {
			t.Fatal("expected invalid short line to be skipped")
		}
	})
}

func TestParseGitLog(t *testing.T) {
	t.Parallel()

	t.Run("populated", func(t *testing.T) {
		t.Parallel()
		input := "abc1234 first commit\ndef5678 second commit\n"
		got := parseGitLog(input)
		want := []string{
			"abc1234: first commit",
			"def5678: second commit",
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

func TestParseDiffStatSummary(t *testing.T) {
	t.Parallel()

	t.Run("populated", func(t *testing.T) {
		t.Parallel()
		input := "internal/service/handlers.go | 8 +++\n internal/service/handlers_test.go | 75 +++++++++++++---------\n 14 files changed, 89 insertions(+), 148 deletions(-)"
		got := parseDiffStatSummary(input)
		want := []string{
			"internal/service/handlers.go | 8 +++",
			"internal/service/handlers_test.go | 75 +++++++++++++---------",
			"14 files changed, 89 insertions(+), 148 deletions(-)",
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("parseDiffStatSummary() = %#v, want %#v", got, want)
		}
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		if len(parseDiffStatSummary("")) != 0 {
			t.Fatal("expected empty slice")
		}
	})
}

func TestGitOverviewWithoutGitRepo(t *testing.T) {
	root := t.TempDir()
	testutil.Chdir(t, root)

	result, err := GitOverview(10)()(root, false)
	if err != nil {
		t.Fatalf("GitOverview() error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil outside git repo, got %+v", result)
	}
}

func TestGitOverviewGitNotInPath(t *testing.T) {
	root := t.TempDir()
	testutil.Chdir(t, root)
	t.Setenv("PATH", t.TempDir())

	result, err := GitOverview(10)()(root, false)
	if err != nil {
		t.Fatalf("GitOverview() error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil when git not in PATH, got %+v", result)
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

	result, err := GitOverview(10)()(root, false)
	if err != nil {
		t.Fatalf("GitOverview() error: %v", err)
	}

	resp, ok := result.(gitOverviewResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if resp.CurrentBranchName == "" {
		t.Fatalf("expected currentBranchName inside git repo, got %+v", resp)
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

	result, err := GitOverview(10)()(root, false)
	if err != nil {
		t.Fatalf("GitOverview() error: %v", err)
	}

	resp, ok := result.(gitOverviewResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if resp.CommitCount == nil || *resp.CommitCount != 1 || resp.SomeRecentCommits == nil || len(*resp.SomeRecentCommits) == 0 {
		t.Fatalf("expected recent commits and commitCount=1, got %+v", resp)
	}
}

func TestGitOverviewEmptyRepoRecentCommits(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not in PATH")
	}

	root := t.TempDir()
	testutil.Chdir(t, root)

	runGit(t, root, "init")
	testutil.WriteFile(t, root+"/README.md", "demo\n")

	result, err := GitOverview(10)()(root, false)
	if err != nil {
		t.Fatalf("GitOverview() error: %v", err)
	}

	resp, ok := result.(gitOverviewResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if resp.CommitCount != nil {
		t.Fatalf("expected commitCount omitted, got %+v", resp.CommitCount)
	}
	if resp.SomeRecentCommits == nil {
		t.Fatal("expected someRecentCommits present when fetch enabled")
	}
	if len(*resp.SomeRecentCommits) != 0 {
		t.Fatalf("expected empty someRecentCommits, got %+v", *resp.SomeRecentCommits)
	}
}

func TestGitOverviewSkipsRecentCommitsWhenZero(t *testing.T) {
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

	result, err := GitOverview(0)()(root, false)
	if err != nil {
		t.Fatalf("GitOverview() error: %v", err)
	}

	resp := result.(gitOverviewResponse)
	if resp.SomeRecentCommits != nil {
		t.Fatalf("expected someRecentCommits omitted when count is 0, got %+v", resp.SomeRecentCommits)
	}
	if resp.CommitCount == nil || *resp.CommitCount != 1 {
		t.Fatalf("expected commitCount=1, got %+v", resp.CommitCount)
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

	result, err := GitOverview(10)()(root, false)
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

func TestGitOverviewCommitCountIsFullHistory(t *testing.T) {
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
	testutil.WriteFile(t, root+"/README.md", "demo2\n")
	runGit(t, root, "add", "README.md")
	runGit(t, root, "commit", "-m", "second")

	result, err := GitOverview(1)()(root, false)
	if err != nil {
		t.Fatalf("GitOverview() error: %v", err)
	}

	resp := result.(gitOverviewResponse)
	if resp.CommitCount == nil || *resp.CommitCount != 2 {
		t.Fatalf("commitCount = %v, want 2", resp.CommitCount)
	}
	if resp.SomeRecentCommits == nil || len(*resp.SomeRecentCommits) != 1 {
		t.Fatalf("someRecentCommits len = %v, want 1", resp.SomeRecentCommits)
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

	result, err := GitOverview(10)()(root, false)
	if err != nil {
		t.Fatalf("GitOverview() error: %v", err)
	}

	resp, ok := result.(gitOverviewResponse)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}

	if resp.DetachedHeadCommitHash == nil || *resp.DetachedHeadCommitHash == "" {
		t.Fatalf("expected detachedHeadCommitHash, got %+v", resp)
	}
	if resp.CommitCount == nil || *resp.CommitCount != 1 {
		t.Fatalf("commitCount = %v, want 1", resp.CommitCount)
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
