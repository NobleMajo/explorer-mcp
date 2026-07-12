package service

import (
	"os/exec"
	"reflect"
	"testing"
)

func TestParseGitStatusShort(t *testing.T) {
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
}

func TestParseGitStatusShortEmpty(t *testing.T) {
	t.Parallel()

	got := parseGitStatusShort("")
	if got == nil || len(got) != 0 {
		t.Fatalf("parseGitStatusShort(\"\") = %#v, want empty slice", got)
	}
}

func TestParseGitLogEmpty(t *testing.T) {
	t.Parallel()

	got := parseGitLog("")
	if got == nil || len(got) != 0 {
		t.Fatalf("parseGitLog(\"\") = %#v, want empty slice", got)
	}
}

func TestGitOverviewWithoutGitRepo(t *testing.T) {
	root := t.TempDir()
	chdir(t, root)

	jsonText, err := GitOverview()
	if err != nil {
		t.Fatalf("GitOverview() error: %v", err)
	}

	var resp gitOverviewResponse
	parseJSONResponse(t, jsonText, &resp)

	if resp.ToolName != "git_overview" {
		t.Fatalf("toolName = %q", resp.ToolName)
	}
	if resp.IsGitRepo {
		t.Fatal("expected isGitRepo false outside git repo")
	}
	if resp.ChangedFiles == nil || resp.RecentCommits == nil {
		t.Fatal("expected non-nil empty slices")
	}
}

func TestGitOverviewInsideGitRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not in PATH")
	}

	root := t.TempDir()
	chdir(t, root)

	initCmd := exec.Command("git", "init")
	initCmd.Dir = root
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}

	writeFile(t, root+"/README.md", "demo\n")

	jsonText, err := GitOverview()
	if err != nil {
		t.Fatalf("GitOverview() error: %v", err)
	}

	var resp gitOverviewResponse
	parseJSONResponse(t, jsonText, &resp)

	if !resp.IsGitAvailable || !resp.IsGitRepo || !resp.IsInsideWorkTree {
		t.Fatalf("unexpected git flags: %+v", resp)
	}
}

func TestParseGitLog(t *testing.T) {
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
}
