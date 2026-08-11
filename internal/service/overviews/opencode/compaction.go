package opencode

import "sort"

// deduplicateRules applies last-wins deduplication keyed on (permission, pattern).
// Preserves first-seen order but uses the action from the last occurrence.
func deduplicateRules(rules []permissionRule) []permissionRule {
	type key struct{ perm, pattern string }
	lastAction := make(map[key]string, len(rules))
	for _, r := range rules {
		lastAction[key{r.Permission, r.Pattern}] = r.Action
	}
	seen := make(map[key]bool, len(lastAction))
	result := make([]permissionRule, 0, len(lastAction))
	for _, r := range rules {
		k := key{r.Permission, r.Pattern}
		if !seen[k] {
			seen[k] = true
			result = append(result, permissionRule{
				Permission: r.Permission,
				Pattern:    r.Pattern,
				Action:     lastAction[k],
			})
		}
	}
	return result
}

// groupRules groups patterns by permission type then by action.
func groupRules(rules []permissionRule) map[string]map[string][]string {
	groups := make(map[string]map[string][]string)
	for _, r := range rules {
		if groups[r.Permission] == nil {
			groups[r.Permission] = make(map[string][]string)
		}
		groups[r.Permission][r.Action] = append(groups[r.Permission][r.Action], r.Pattern)
	}
	return groups
}

// compactPatterns returns a single glob string when there is only one pattern,
// otherwise a sorted list of globs.
func compactPatterns(patterns []string) any {
	sorted := append([]string(nil), patterns...)
	sort.Strings(sorted)
	if len(sorted) == 1 {
		return sorted[0]
	}
	return sorted
}

// compactPermission collapses a permission's action→patterns map:
//   - single action with only "*"  → just the action string ("ask")
//   - otherwise                    → map[action]string|[]string of globs
func compactPermission(actionMap map[string][]string) any {
	if len(actionMap) == 1 {
		for action, patterns := range actionMap {
			if len(patterns) == 1 && patterns[0] == "*" {
				return action
			}
		}
	}
	built := make(map[string]any, len(actionMap))
	for action, patterns := range actionMap {
		built[action] = compactPatterns(patterns)
	}
	return built
}

// compactPermissions deduplicates rules, groups by permission+action, and keeps
// the original glob patterns (sorted). Trivial single-action "*" permissions
// collapse to just the action string. Returns nil for empty input.
//
// Example shape:
//
//	{
//	  "*": "ask",
//	  "websearch": "ask",
//	  "bash": {
//	    "allow": ["cat *", "docker logs *", "git log *"],
//	    "ask": ["*", "docker *"],
//	    "deny": ["bun *", "cargo *"]
//	  },
//	  "read": {
//	    "allow": "*",
//	    "ask": ["*.env", "*.env.*"],
//	    "deny": ["*.env*", "*.netrc*"]
//	  }
//	}
func compactPermissions(rules []permissionRule) map[string]any {
	if len(rules) == 0 {
		return nil
	}
	groups := groupRules(deduplicateRules(rules))
	result := make(map[string]any, len(groups))
	for perm, actionMap := range groups {
		result[perm] = compactPermission(actionMap)
	}
	return result
}
