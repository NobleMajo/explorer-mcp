package globals

import (
	"reflect"
	"testing"
)

func TestCollectProjectIdentifierFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		subfiles []string
		want     []string
	}{
		{
			name:     "makefile",
			subfiles: []string{"Makefile", "README.md"},
			want:     []string{"@makefile"},
		},
		{
			name:     "tsconfig json",
			subfiles: []string{"tsconfig.json"},
			want:     []string{"@tsconfig"},
		},
		{
			name:     "prefixed tsconfig",
			subfiles: []string{"app.tsconfig.json"},
			want:     []string{"@tsconfig"},
		},
		{
			name:     "tsconfig env file",
			subfiles: []string{"tsconfig.app.json"},
			want:     []string{"@tsconfig"},
		},
		{
			name:     "angular",
			subfiles: []string{"angular.json"},
			want:     []string{"@angular"},
		},
		{
			name:     "combined",
			subfiles: []string{"Makefile", "angular.json", "tsconfig.json"},
			want:     []string{"@makefile", "@tsconfig", "@angular"},
		},
		{
			name:     "none",
			subfiles: []string{"README.md"},
			want:     nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := CollectProjectIdentifierFlags("/tmp/project", tc.subfiles, nil)
			if err != nil {
				t.Fatalf("CollectProjectIdentifierFlags() error: %v", err)
			}
			if len(tc.want) == 0 {
				if len(got) != 0 {
					t.Fatalf("CollectProjectIdentifierFlags() = %#v, want empty", got)
				}
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("CollectProjectIdentifierFlags() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestIsTSConfigFileName(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"tsconfig.json", "app.tsconfig.json", "tsconfig.app.json"} {
		if !isTSConfigFileName(name) {
			t.Fatalf("isTSConfigFileName(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"package.json", "config.json", "nontsconfig.app.json"} {
		if isTSConfigFileName(name) {
			t.Fatalf("isTSConfigFileName(%q) = true, want false", name)
		}
	}
}
