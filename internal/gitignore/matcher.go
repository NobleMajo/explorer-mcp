package gitignore

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

const gitIgnoreFileName = ".gitignore"

type DirMatcher struct {
	baseRel string
	rules   []rule
}

type rule struct {
	negated  bool
	dirOnly  bool
	pattern  string
	anchored bool
}

func LoadDirMatcher(dir, root string) (*DirMatcher, error) {
	path := filepath.Join(dir, gitIgnoreFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	baseRel, err := filepath.Rel(root, dir)
	if err != nil {
		return nil, err
	}
	baseRel = filepath.ToSlash(baseRel)
	if baseRel == "." {
		baseRel = ""
	}

	return &DirMatcher{
		baseRel: baseRel,
		rules:   parseRules(string(data)),
	}, nil
}

func parseRules(content string) []rule {
	rules := make([]rule, 0)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		negated := false
		if strings.HasPrefix(line, "!") {
			negated = true
			line = strings.TrimSpace(line[1:])
			if line == "" {
				continue
			}
		}

		dirOnly := strings.HasSuffix(line, "/")
		line = strings.TrimSuffix(line, "/")

		anchored := strings.HasPrefix(line, "/")
		if anchored {
			line = strings.TrimPrefix(line, "/")
		}

		if line == "" {
			continue
		}

		rules = append(rules, rule{
			negated:  negated,
			dirOnly:  dirOnly,
			pattern:  line,
			anchored: anchored,
		})
	}

	return rules
}

func ShouldIgnore(matchers []DirMatcher, relPath string, isDir bool) bool {
	path := filepath.ToSlash(relPath)
	if path == "." {
		return false
	}

	ignored := false
	matched := false
	for _, matcher := range matchers {
		relToMatcher := path
		if matcher.baseRel != "" {
			if path == matcher.baseRel {
				relToMatcher = ""
			} else if strings.HasPrefix(path, matcher.baseRel+"/") {
				relToMatcher = strings.TrimPrefix(path, matcher.baseRel+"/")
			} else {
				continue
			}
		}

		if ruleMatched, ruleIgnored := matcher.match(relToMatcher, isDir); ruleMatched {
			matched = true
			ignored = ruleIgnored
		}
	}

	return matched && ignored
}

func (m DirMatcher) match(relPath string, isDir bool) (matched bool, ignored bool) {
	name := filepath.Base(relPath)
	for _, pattern := range m.rules {
		if pattern.dirOnly && !isDir {
			continue
		}
		if !pattern.matches(relPath, name) {
			continue
		}
		matched = true
		ignored = !pattern.negated
	}
	return matched, ignored
}

func (r rule) matches(relPath, baseName string) bool {
	if r.anchored {
		return matchPattern(r.pattern, relPath)
	}
	if matchPattern(r.pattern, baseName) {
		return true
	}
	if strings.Contains(relPath, "/") {
		return matchPattern(r.pattern, relPath)
	}
	return false
}

func matchPattern(pattern, value string) bool {
	if pattern == "" {
		return value == ""
	}

	parts := strings.Split(pattern, "**")
	if len(parts) > 1 {
		return matchDoubleStar(parts, value)
	}

	return matchSegments(strings.Split(pattern, "/"), strings.Split(value, "/"))
}

func matchDoubleStar(parts []string, value string) bool {
	if len(parts) == 0 {
		return true
	}

	segments := strings.Split(value, "/")
	return matchDoubleStarParts(parts, segments, 0, 0)
}

func matchDoubleStarParts(parts []string, segments []string, partIdx, segIdx int) bool {
	if partIdx == len(parts) {
		return segIdx == len(segments)
	}

	current := strings.Trim(parts[partIdx], "/")
	isLast := partIdx == len(parts)-1

	if current == "" {
		if isLast {
			return true
		}
		for start := segIdx; start <= len(segments); start++ {
			if matchDoubleStarParts(parts, segments, partIdx+1, start) {
				return true
			}
		}
		return false
	}

	subparts := strings.Split(current, "/")
	for start := segIdx; start <= len(segments); start++ {
		if !matchSegments(subparts, segments[segIdx:start]) {
			continue
		}
		next := start
		if isLast {
			if next == len(segments) {
				return true
			}
			continue
		}
		if matchDoubleStarParts(parts, segments, partIdx+1, next) {
			return true
		}
	}
	return false
}

func matchSegments(pattern, value []string) bool {
	if len(pattern) == 0 {
		return len(value) == 0
	}
	if len(value) == 0 {
		for _, part := range pattern {
			if part != "" && part != "**" {
				return false
			}
		}
		return true
	}
	if len(pattern) != len(value) {
		return false
	}
	for i := range pattern {
		if !matchSimple(pattern[i], value[i]) {
			return false
		}
	}
	return true
}

func matchSimple(pattern, value string) bool {
	if pattern == "" {
		return value == ""
	}

	pi, vi := 0, 0
	for pi < len(pattern) {
		if pattern[pi] == '*' {
			if pi == len(pattern)-1 {
				return true
			}
			for end := vi; end <= len(value); end++ {
				if matchSimple(pattern[pi+1:], value[end:]) {
					return true
				}
			}
			return false
		}
		if vi >= len(value) {
			return false
		}
		if pattern[pi] == '?' {
			pi++
			vi++
			continue
		}
		if pattern[pi] != value[vi] {
			return false
		}
		pi++
		vi++
	}
	return vi == len(value)
}
