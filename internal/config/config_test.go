package config

import (
	"os"
	"testing"
)

func TestParseConfigPrintCommand(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	os.Args = []string{"explorer-mcp", "print"}
	cfg := ParseConfig("Demo", "demo", "1.0.0", "abc")
	if !cfg.PrintAll {
		t.Fatal("expected PrintAll true")
	}
}
