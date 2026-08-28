package workspacescan

import "testing"

func TestIsOutputDirAndIsDepsDir(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"dist", "out", "output"} {
		if !IsOutputDir(name) {
			t.Fatalf("expected %q to be output dir", name)
		}
	}
	for _, name := range []string{"node_modules", "vendor"} {
		if !IsDepsDir(name) {
			t.Fatalf("expected %q to be deps dir", name)
		}
	}
	for _, name := range []string{"build", "dist-extra", "myout", "src", "node_module"} {
		if IsOutputDir(name) {
			t.Fatalf("expected %q not to be output dir", name)
		}
		if IsDepsDir(name) {
			t.Fatalf("expected %q not to be deps dir", name)
		}
	}
}

func TestShouldCollapseDir(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		opts          Options
		dirName       string
		wantCollapsed bool
	}{
		{
			name:          "output dir collapsed by default",
			opts:          Options{},
			dirName:       "dist",
			wantCollapsed: true,
		},
		{
			name:          "output dir expanded when enabled",
			opts:          Options{ExpandOutDirs: true},
			dirName:       "dist",
			wantCollapsed: false,
		},
		{
			name:          "deps dir collapsed by default",
			opts:          Options{},
			dirName:       "node_modules",
			wantCollapsed: true,
		},
		{
			name:          "deps dir expanded when enabled",
			opts:          Options{ExpandDepsDirs: true},
			dirName:       "vendor",
			wantCollapsed: false,
		},
		{
			name:          "regular dir not collapsed",
			opts:          Options{},
			dirName:       "src",
			wantCollapsed: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.opts.shouldCollapseDir(tc.dirName)
			if got != tc.wantCollapsed {
				t.Fatalf("shouldCollapseDir(%q) = %v, want %v", tc.dirName, got, tc.wantCollapsed)
			}
		})
	}
}

func TestTruncationSuffix(t *testing.T) {
	t.Parallel()

	if got := (Options{}).truncationSuffix(); got != "/**" {
		t.Fatalf("default truncation suffix = %q, want /**", got)
	}
	if got := (Options{TruncateMarker: "/..."}).truncationSuffix(); got != "/..." {
		t.Fatalf("custom truncation suffix = %q, want /...", got)
	}
}
