package git

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type gitOverviewResponse struct {
	IsGitAvailable          bool     `json:"isGitAvailable"`
	IsGitRepo               bool     `json:"isGitRepo"`
	IsInsideWorkTree        bool     `json:"isInsideWorkTree"`
	CurrentBranchName       string   `json:"currentBranchName"`
	IsDetachedHead          bool     `json:"isDetachedHead"`
	DetachedHeadCommitHash  string   `json:"detachedHeadCommitHash"`
	IsWorkingTreeClean      bool     `json:"isWorkingTreeClean"`
	ChangedFileCount        int      `json:"changedFileCount"`
	ChangedFiles            []string `json:"changedFiles"`
	RecentCommitsListed     bool     `json:"recentCommitsListed"`
	CommitCount             *int     `json:"commitCount,omitempty"`
	SomeRecentCommits       []string `json:"someRecentCommits,omitempty"`
	UnstagedDiffStatSummary []string `json:"unstagedDiffStatSummary"`
	ErrorMessage            string   `json:"errorMessage,omitempty"`
}

func buildGitOverview(verbose bool, recentCommitCount int) (any, error) {
	_ = verbose
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	if _, err := exec.LookPath("git"); err != nil {
		return nil, nil
	}

	inside, err := gitOutput(dir, "rev-parse", "--is-inside-work-tree")
	if err != nil || inside != "true" {
		return nil, nil
	}

	resp := gitOverviewResponse{
		IsGitAvailable:          true,
		ChangedFiles:            []string{},
		UnstagedDiffStatSummary: []string{},
		RecentCommitsListed:     recentCommitCount > 0,
		IsGitRepo:               true,
		IsInsideWorkTree:        true,
	}

	if count, ok := gitHistoryCommitCount(dir); ok {
		resp.CommitCount = &count
	}

	branch, _ := gitOutput(dir, "branch", "--show-current")
	resp.CurrentBranchName = branch
	if branch == "" {
		resp.IsDetachedHead = true
		resp.DetachedHeadCommitHash, _ = gitOutput(dir, "rev-parse", "--short", "HEAD")
	}

	statusOut, _ := gitOutput(dir, "status", "--short")
	resp.ChangedFiles = parseGitStatusShort(statusOut)
	resp.ChangedFileCount = len(resp.ChangedFiles)
	resp.IsWorkingTreeClean = resp.ChangedFileCount == 0

	if recentCommitCount > 0 {
		logOut, _ := gitOutput(dir, "log", fmt.Sprintf("-%d", recentCommitCount), "--format=%h %s")
		commits := parseGitLog(logOut)
		if len(commits) > 0 {
			resp.SomeRecentCommits = commits
		}
	}

	statOut, _ := gitOutput(dir, "diff", "--stat")
	resp.UnstagedDiffStatSummary = parseDiffStatSummary(statOut)

	return resp, nil
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func gitHistoryCommitCount(dir string) (int, bool) {
	out, err := gitOutput(dir, "rev-list", "--count", "HEAD")
	if err != nil {
		return 0, false
	}

	count, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return 0, false
	}
	return count, true
}

func parseGitStatusShort(output string) []string {
	if output == "" {
		return []string{}
	}

	lines := strings.Split(output, "\n")
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimRight(line, " \t")
		if len(line) < 3 {
			continue
		}

		filePath := strings.TrimSpace(line[2:])
		if filePath == "" {
			continue
		}

		status := compactGitStatus(line[0], line[1])
		files = append(files, status+": "+filePath)
	}
	return files
}

func compactGitStatus(staged, worktree byte) string {
	switch {
	case staged == '?' && worktree == '?':
		return "?"
	case staged == '!' && worktree == '!':
		return "!"
	case staged == 'M' && worktree == 'M':
		return "*"
	case staged == 'M' && worktree == ' ':
		return "M"
	case staged == ' ' && worktree == 'M':
		return "m"
	case staged == 'A' && (worktree == ' ' || worktree == 'M'):
		return "+"
	case staged == 'D' && worktree == ' ':
		return "D"
	case staged == ' ' && worktree == 'D':
		return "d"
	case staged == 'D' && worktree == 'D':
		return "D"
	case staged == 'R' || worktree == 'R':
		return "r"
	case staged == 'C' || worktree == 'C':
		return "c"
	case staged == 'U' || worktree == 'U':
		return "u"
	default:
		code := strings.TrimSpace(string([]byte{staged, worktree}))
		if code == "" {
			return "."
		}
		return code
	}
}

func parseDiffStatSummary(output string) []string {
	if output == "" {
		return []string{}
	}

	lines := strings.Split(output, "\n")
	summary := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		summary = append(summary, line)
	}
	return summary
}

func parseGitLog(output string) []string {
	if output == "" {
		return []string{}
	}

	lines := strings.Split(output, "\n")
	commits := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		hash, subject, ok := strings.Cut(line, " ")
		if !ok || hash == "" || subject == "" {
			continue
		}
		commits = append(commits, hash+": "+subject)
	}
	return commits
}
