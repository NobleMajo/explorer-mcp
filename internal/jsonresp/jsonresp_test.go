package jsonresp

import (
	"encoding/json"
	"testing"
)

func TestMarshal(t *testing.T) {
	t.Parallel()

	got, err := Marshal(Meta{
		ToolName:      "demo",
		SchemaVersion: SchemaVersion,
	})
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var meta Meta
	if err := json.Unmarshal([]byte(got), &meta); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if meta.ToolName != "demo" || meta.SchemaVersion != SchemaVersion {
		t.Fatalf("unexpected meta: %+v", meta)
	}
}

func TestMarshalError(t *testing.T) {
	t.Parallel()

	_, err := Marshal(make(chan int))
	if err == nil {
		t.Fatal("expected marshal error")
	}
}
