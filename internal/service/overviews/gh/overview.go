package gh

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	cliName               = "gh"
	repoListLimit         = 1000
	noSimilarReposMessage = "no similar user or org repo found"
)

type ghRepo struct {
	Name        string `json:"name"`
	ID          any    `json:"id"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

// Overridable in tests.
var (
	lookPath = exec.LookPath
	runGh    = func(dir string, args ...string) (string, error) {
		cmd := exec.Command(cliName, args...)
		cmd.Dir = dir
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(out)), nil
	}
	runGit = func(dir string, args ...string) (string, error) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(out)), nil
	}
)

func buildGhOverview(projectRootPath string, verbose bool) (any, error) {
	_ = verbose

	if _, err := lookPath(cliName); err != nil {
		return nil, nil
	}

	if _, err := runGh(projectRootPath, "auth", "status"); err != nil {
		return nil, nil
	}

	orgs, err := listOrgs(projectRootPath)
	if err != nil {
		orgs = nil
	}

	repos := collectRepos(projectRootPath, orgs)
	needles := collectNeedles(projectRootPath)
	matches := filterMatchingRepos(repos, needles)
	if len(matches) == 0 {
		return []string{noSimilarReposMessage}, nil
	}

	lines := make([]string, 0, len(matches))
	for _, repo := range matches {
		lines = append(lines, formatRepoLine(repo))
	}
	return lines, nil
}

func listOrgs(projectRootPath string) ([]string, error) {
	raw, err := orgListCache.getOrFetch("org-list", orgListCacheTTL, func() (string, error) {
		return runGh(projectRootPath, "org", "list")
	})
	if err != nil {
		return nil, err
	}
	return formatOrgLines(raw), nil
}

func formatOrgLines(raw string) []string {
	if raw == "" {
		return []string{}
	}

	lines := strings.Split(raw, "\n")
	orgs := make([]string, 0, len(lines))
	for _, line := range lines {
		org := strings.TrimSpace(line)
		if org == "" {
			continue
		}
		orgs = append(orgs, org)
	}
	return orgs
}

func collectRepos(projectRootPath string, orgs []string) []ghRepo {
	combined := make([]ghRepo, 0)
	seen := make(map[string]struct{})

	appendRepos := func(raw string) {
		for _, repo := range parseRepoListJSON(raw) {
			id := repoIDString(repo.ID)
			if id != "" {
				if _, ok := seen[id]; ok {
					continue
				}
				seen[id] = struct{}{}
			}
			combined = append(combined, repo)
		}
	}

	if raw, err := repoListCache.getOrFetch("repos:user", repoListCacheTTL, func() (string, error) {
		return runGh(projectRootPath, "repo", "list", "--limit", fmt.Sprintf("%d", repoListLimit), "--json", "name,id,url,description")
	}); err == nil {
		appendRepos(raw)
	}

	for _, org := range orgs {
		org := org
		key := "repos:org:" + org
		raw, err := repoListCache.getOrFetch(key, repoListCacheTTL, func() (string, error) {
			return runGh(projectRootPath, "repo", "list", org, "--limit", fmt.Sprintf("%d", repoListLimit), "--json", "name,id,url,description")
		})
		if err != nil {
			continue
		}
		appendRepos(raw)
	}

	return combined
}

func parseRepoListJSON(raw string) []ghRepo {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var repos []ghRepo
	if err := json.Unmarshal([]byte(raw), &repos); err != nil {
		return nil
	}
	return repos
}

func collectNeedles(projectRootPath string) []string {
	needles := make([]string, 0, 4)
	seen := make(map[string]struct{})

	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		needles = append(needles, value)
	}

	add(filepath.Base(projectRootPath))

	for _, name := range remoteRepoNames(projectRootPath) {
		add(name)
	}

	return needles
}

func remoteRepoNames(projectRootPath string) []string {
	out, err := runGit(projectRootPath, "remote", "-v")
	if err != nil || out == "" {
		return nil
	}

	names := make([]string, 0)
	seen := make(map[string]struct{})
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := repoNameFromRemoteURL(fields[1])
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		names = append(names, name)
	}
	return names
}

func repoNameFromRemoteURL(rawURL string) string {
	_, name := ownerAndNameFromURL(rawURL)
	return name
}

func ownerAndNameFromURL(rawURL string) (owner, name string) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", ""
	}

	rawURL = strings.TrimSuffix(rawURL, ".git")

	// git@host:owner/repo or ssh://git@host/owner/repo
	if strings.HasPrefix(rawURL, "git@") {
		_, path, ok := strings.Cut(rawURL, ":")
		if !ok {
			return "", ""
		}
		return splitOwnerName(path)
	}

	if i := strings.Index(rawURL, "://"); i >= 0 {
		rest := rawURL[i+3:]
		parts := strings.Split(rest, "/")
		if len(parts) < 3 {
			return "", ""
		}
		return parts[1], parts[2]
	}

	// host:owner/repo without scheme
	if _, path, ok := strings.Cut(rawURL, ":"); ok && !strings.Contains(rawURL, "/") {
		return splitOwnerName(path)
	}

	return splitOwnerName(rawURL)
}

func splitOwnerName(path string) (owner, name string) {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		if len(parts) == 1 {
			return "", parts[0]
		}
		return "", ""
	}
	return parts[len(parts)-2], parts[len(parts)-1]
}

func filterMatchingRepos(repos []ghRepo, needles []string) []ghRepo {
	if len(needles) == 0 {
		return nil
	}

	lowerNeedles := make([]string, len(needles))
	for i, needle := range needles {
		lowerNeedles[i] = strings.ToLower(needle)
	}

	matches := make([]ghRepo, 0)
	for _, repo := range repos {
		haystack := strings.ToLower(strings.Join([]string{
			repo.Name,
			repoIDString(repo.ID),
			repo.URL,
			repo.Description,
		}, "\n"))
		for _, needle := range lowerNeedles {
			if needle != "" && strings.Contains(haystack, needle) {
				matches = append(matches, repo)
				break
			}
		}
	}
	return matches
}

func formatRepoLine(repo ghRepo) string {
	owner, name := ownerAndNameFromURL(repo.URL)
	if name == "" {
		name = repo.Name
	}
	if owner == "" {
		owner = "unknown"
	}
	if name == "" {
		name = "unknown"
	}
	return fmt.Sprintf("%s/%s @%s @%s", owner, name, repoIDString(repo.ID), repo.URL)
}

func repoIDString(id any) string {
	if id == nil {
		return ""
	}
	switch v := id.(type) {
	case string:
		return v
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%v", v)
	case json.Number:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}
