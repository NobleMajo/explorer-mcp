package globals

import "testing"

func TestIsScanIgnored(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "git dir", input: ".git", expected: true},
		{name: "__pycache__", input: "__pycache__", expected: true},
		{name: "source file", input: "main.go", expected: false},
		{name: "readme", input: "README.md", expected: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := IsScanIgnored(tc.input); got != tc.expected {
				t.Fatalf("IsScanIgnored(%q) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}
