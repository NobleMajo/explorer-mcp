package git

import (
	"os"
	"os/exec"
	"strings"

	"github.com/NobleMajo/explorer-mcp/internal/jsonresp"
)

type gitOverviewResponse struct {
	jsonresp.Meta
	IsGitAvailable          bool             `json:"isGitAvailable"`
	IsGitRepo               bool             `json:"isGitRepo"`
	IsInsideWorkTree        bool             `json:"isInsideWorkTree"`
	CurrentBranchName       string           `json:"currentBranchName"`
	IsDetachedHead          bool             `json:"isDetachedHead"`
	DetachedHeadCommitHash  string           `json:"detachedHeadCommitHash"`
	IsWorkingTreeClean      bool             `json:"isWorkingTreeClean"`
	ChangedFileCount        int              `json:"changedFileCount"`
	ChangedFiles            []gitChangedFile `json:"changedFiles"`
	RecentCommitCount       int              `json:"recentCommitCount"`
	RecentCommits           []gitCommit      `json:"recentCommits"`
	UnstagedDiffStatSummary string           `json:"unstagedDiffStatSummary"`
	ErrorMessage            string           `json:"errorMessage,omitempty"`
}

type gitChangedFile struct {
	StatusCode string `json:"statusCode"`
	FilePath   string `json:"filePath"`
}

type gitCommit struct {
	ShortCommitHash string `json:"shortCommitHash"`
	CommitSubject   string `json:"commitSubject"`
}

func buildGitOverview(verbose bool) (gitOverviewResponse, error) {
	_ = verbose
	dir, err := os.Getwd()
	if err != nil {
		return gitOverviewResponse{}, err
	}

	resp := gitOverviewResponse{
		Meta: jsonresp.Meta{
			ToolName:      "git_overview",
			SchemaVersion: jsonresp.SchemaVersion,
		},
		ChangedFiles:  []gitChangedFile{},
		RecentCommits: []gitCommit{},
	}

	if _, err := exec.LookPath("git"); err != nil {
		resp.IsGitAvailable = false
		resp.ErrorMessage = "git executable not found in PATH"
		return resp, nil
	}
	resp.IsGitAvailable = true

	inside, err := gitOutput(dir, "rev-parse", "--is-inside-work-tree")
	if err != nil || inside != "true" {
		return resp, nil
	}

	resp.IsGitRepo = true
	resp.IsInsideWorkTree = true

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

	logOut, _ := gitOutput(dir, "log", "-5", "--format=%h %s")
	resp.RecentCommits = parseGitLog(logOut)
	resp.RecentCommitCount = len(resp.RecentCommits)

	resp.UnstagedDiffStatSummary, _ = gitOutput(dir, "diff", "--stat")

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

func parseGitStatusShort(output string) []gitChangedFile {
	if output == "" {
		return []gitChangedFile{}
	}

	lines := strings.Split(output, "\n")
	files := make([]gitChangedFile, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimRight(line, " \t")
		if line == "" {
			continue
		}

		statusCode := line
		filePath := ""
		if len(line) >= 3 {
			statusCode = line[:2]
			filePath = strings.TrimSpace(line[2:])
		}

		files = append(files, gitChangedFile{
			StatusCode: statusCode,
			FilePath:   filePath,
		})
	}
	return files
}

func parseGitLog(output string) []gitCommit {
	if output == "" {
		return []gitCommit{}
	}

	lines := strings.Split(output, "\n")
	commits := make([]gitCommit, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		hash, subject, _ := strings.Cut(line, " ")
		commits = append(commits, gitCommit{
			ShortCommitHash: hash,
			CommitSubject:   subject,
		})
	}
	return commits
}
