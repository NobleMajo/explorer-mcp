package gitignore

import (
	"path/filepath"
	"testing"

	"github.com/NobleMajo/explorer-mcp/internal/testutil"
)

func TestParseRulesSkipsCommentsAndBlanks(t *testing.T) {
	t.Parallel()

	rules := parseRules("# comment\n\n*.log\n")
	if len(rules) != 1 || rules[0].pattern != "*.log" {
		t.Fatalf("unexpected rules: %+v", rules)
	}
}

func TestShouldIgnoreWildcardFile(t *testing.T) {
	t.Parallel()

	matchers := []DirMatcher{{
		rules: parseRules("*.log\n"),
	}}

	if !ShouldIgnore(matchers, "debug.log", false) {
		t.Fatal("expected debug.log to be ignored")
	}
	if ShouldIgnore(matchers, "main.go", false) {
		t.Fatal("expected main.go to be included")
	}
}

func TestShouldIgnoreDirectoryPattern(t *testing.T) {
	t.Parallel()

	matchers := []DirMatcher{{
		rules: parseRules("dist/\n"),
	}}

	if !ShouldIgnore(matchers, "dist", true) {
		t.Fatal("expected dist directory to be ignored")
	}
	if ShouldIgnore(matchers, "dist/output.txt", false) {
		t.Fatal("expected file under dist to be included when only parent dir is skipped")
	}
}

func TestShouldIgnoreAnchoredPattern(t *testing.T) {
	t.Parallel()

	matchers := []DirMatcher{{
		rules: parseRules("/build\n"),
	}}

	if !ShouldIgnore(matchers, "build", true) {
		t.Fatal("expected root build directory to be ignored")
	}
	if ShouldIgnore(matchers, "pkg/build", true) {
		t.Fatal("expected nested build directory to be included")
	}
}

func TestShouldIgnoreNestedGitignore(t *testing.T) {
	t.Parallel()

	matchers := []DirMatcher{
		{baseRel: "", rules: parseRules("tmp/\n")},
		{baseRel: "pkg", rules: parseRules("generated/\n")},
	}

	if !ShouldIgnore(matchers, "pkg/generated", true) {
		t.Fatal("expected nested gitignore directory to be ignored")
	}
	if ShouldIgnore(matchers, "pkg/manual", true) {
		t.Fatal("expected unrelated pkg directory to be included")
	}
}

func TestShouldIgnoreDoubleStarPattern(t *testing.T) {
	t.Parallel()

	matchers := []DirMatcher{{
		rules: parseRules("logs/**\n"),
	}}

	if !ShouldIgnore(matchers, "logs/app/error.log", false) {
		t.Fatal("expected nested file under logs to be ignored")
	}
	if ShouldIgnore(matchers, "other/error.log", false) {
		t.Fatal("expected file outside logs to be included")
	}
}

func TestShouldIgnorePathPattern(t *testing.T) {
	t.Parallel()

	matchers := []DirMatcher{{
		rules: parseRules("src/?ebug.log\n"),
	}}

	if !ShouldIgnore(matchers, "src/debug.log", false) {
		t.Fatal("expected src/debug.log to be ignored")
	}
	if ShouldIgnore(matchers, "src/trace.log", false) {
		t.Fatal("expected src/trace.log to be included")
	}
}

func TestParseRulesSkipsEmptyNegation(t *testing.T) {
	t.Parallel()

	rules := parseRules("!\n*.log\n")
	if len(rules) != 1 || rules[0].pattern != "*.log" {
		t.Fatalf("unexpected rules: %+v", rules)
	}
}

func TestShouldIgnoreNegation(t *testing.T) {
	t.Parallel()

	matchers := []DirMatcher{{
		rules: parseRules("*.log\n!important.log\n"),
	}}

	if !ShouldIgnore(matchers, "debug.log", false) {
		t.Fatal("expected debug.log to be ignored")
	}
	if ShouldIgnore(matchers, "important.log", false) {
		t.Fatal("expected important.log to be un-ignored")
	}
}

func TestLoadDirMatcherMissingFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	matcher, err := LoadDirMatcher(root, root)
	if err != nil {
		t.Fatalf("LoadDirMatcher() error: %v", err)
	}
	if matcher != nil {
		t.Fatal("expected nil matcher for missing .gitignore")
	}
}

func TestLoadDirMatcherReadsFile(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, ".gitignore"), "dist/\n")

	matcher, err := LoadDirMatcher(root, root)
	if err != nil {
		t.Fatalf("LoadDirMatcher() error: %v", err)
	}
	if matcher == nil || !ShouldIgnore([]DirMatcher{*matcher}, "dist", true) {
		t.Fatalf("unexpected matcher: %+v", matcher)
	}
}
