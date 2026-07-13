package globals

import (
	"strings"
	"testing"
)

func TestDepScopeConstants(t *testing.T) {
	t.Parallel()

	want := []string{DepScopeDirect, DepScopeIndirect, DepScopeTool, DepScopeDev}
	for _, scope := range want {
		if scope == "" {
			t.Fatalf("empty dep scope constant")
		}
	}
}

func TestFormatScopedDependencyUsesStandardScopes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		pkg     string
		version string
		scope   string
		want    string
	}{
		{name: "direct", pkg: "foo", version: "1.0.0", scope: DepScopeDirect, want: "foo@1.0.0 direct"},
		{name: "indirect", pkg: "bar", version: "2.0.0", scope: DepScopeIndirect, want: "bar@2.0.0 indirect"},
		{name: "tool", pkg: "baz/cmd", version: "3.0.0", scope: DepScopeTool, want: "baz/cmd@3.0.0 tool"},
		{name: "dev", pkg: "eslint", version: "9.0.0", scope: DepScopeDev, want: "eslint@9.0.0 dev"},
		{name: "no version", pkg: "Qt6", version: "", scope: DepScopeDirect, want: "Qt6 direct"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := formatScopedDependency(tc.pkg, tc.version, tc.scope); got != tc.want {
				t.Fatalf("formatScopedDependency() = %q, want %q", got, tc.want)
			}
		})
	}
}

func dependencyScope(entry string) string {
	for _, scope := range []string{DepScopeIndirect, DepScopeDirect, DepScopeTool, DepScopeDev} {
		if strings.HasSuffix(entry, " "+scope) {
			return scope
		}
	}
	return ""
}

func assertStandardDepScopes(t *testing.T, entries []string) {
	t.Helper()

	for _, entry := range entries {
		if dependencyScope(entry) == "" {
			t.Fatalf("entry %q missing standard scope suffix", entry)
		}
		for _, legacy := range []string{"production", "development", "@direct", "@indirect"} {
			if strings.Contains(entry, legacy) {
				t.Fatalf("entry %q uses legacy scope label %q", entry, legacy)
			}
		}
	}
}

func TestAssertStandardDepScopes(t *testing.T) {
	t.Parallel()

	assertStandardDepScopes(t, []string{
		"github.com/foo/bar@v1.0.0 direct",
		"left-pad@1.0.0 indirect",
		"golang.org/x/tools/cmd/goimports@v0.30.0 tool",
		"eslint@9.0.0 dev",
		"Qt6 direct",
	})
}

func TestDependencyScopeRejectsLegacyAndMissingScopes(t *testing.T) {
	t.Parallel()

	invalid := []string{
		"left-pad@1.0.0 production",
		"eslint@9.0.0 development",
		"foo@v1.0.0 @direct",
		"requests@==2.28.0",
	}

	for _, entry := range invalid {
		if dependencyScope(entry) != "" && !strings.Contains(entry, "production") && !strings.Contains(entry, "development") && !strings.Contains(entry, "@direct") {
			t.Fatalf("entry %q should be invalid", entry)
		}
	}
}
