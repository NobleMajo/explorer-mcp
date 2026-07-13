package config

import (
	"os"
	"testing"
)

func TestParseConfigDirectOutFlag(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	os.Args = []string{"explorer-mcp", "--out"}
	cfg := ParseConfig("Demo", "demo", "1.0.0", "abc")
	if !cfg.DirectOut {
		t.Fatal("expected DirectOut true")
	}
}
