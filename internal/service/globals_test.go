package service

import "testing"

func TestIsScanIgnored(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "git dir", input: ".git", expected: true},
		{name: "node_modules", input: "node_modules", expected: true},
		{name: "source file", input: "main.go", expected: false},
		{name: "readme", input: "README.md", expected: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isScanIgnored(tc.input); got != tc.expected {
				t.Fatalf("isScanIgnored(%q) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestIsIgnoredFile(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "gitignore", input: ".gitignore", expected: true},
		{name: "dockerignore", input: ".dockerignore", expected: true},
		{name: "go mod", input: "go.mod", expected: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isIgnoredFile(tc.input); got != tc.expected {
				t.Fatalf("isIgnoredFile(%q) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}
