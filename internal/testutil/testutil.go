package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func Chdir(t *testing.T, dir string) {
	t.Helper()

	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(old)
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
}

func WriteFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func ParseJSON(t *testing.T, jsonText string, dest any) {
	t.Helper()

	if err := json.Unmarshal([]byte(jsonText), dest); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
}
