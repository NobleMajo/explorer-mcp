package service

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestMarshalResponse(t *testing.T) {
	t.Parallel()

	got, err := marshalResponse(responseMeta{
		ToolName:      "demo",
		SchemaVersion: schemaVersion,
	})
	if err != nil {
		t.Fatalf("marshalResponse() error: %v", err)
	}

	var meta responseMeta
	if err := json.Unmarshal([]byte(got), &meta); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if meta.ToolName != "demo" || meta.SchemaVersion != schemaVersion {
		t.Fatalf("unexpected meta: %+v", meta)
	}
}

func TestLimitJSONOutputUnderCap(t *testing.T) {
	t.Parallel()

	input := `{"toolName":"repo_structure","schemaVersion":1,"entryCount":1}`
	got, err := limitJSONOutput(input)
	if err != nil {
		t.Fatalf("limitJSONOutput() error: %v", err)
	}
	if got != input {
		t.Fatalf("limitJSONOutput() changed small payload")
	}
}

func TestLimitJSONOutputOverCap(t *testing.T) {
	t.Parallel()

	input := `{"toolName":"repo_structure","schemaVersion":1,"payload":"` + strings.Repeat("x", maxToolJSONBytes) + `"}`

	got, err := limitJSONOutput(input)
	if err != nil {
		t.Fatalf("limitJSONOutput() error: %v", err)
	}

	var resp truncatedToolResponse
	if err := json.Unmarshal([]byte(got), &resp); err != nil {
		t.Fatalf("unmarshal truncated response: %v", err)
	}

	if !resp.IsOutputTruncated {
		t.Fatal("expected isOutputTruncated true")
	}
	if resp.ToolName != "repo_structure" {
		t.Fatalf("toolName = %q, want repo_structure", resp.ToolName)
	}
	if resp.OutputByteCount <= maxToolJSONBytes {
		t.Fatalf("outputByteCount = %d, want > %d", resp.OutputByteCount, maxToolJSONBytes)
	}
	if len(resp.TruncatedJsonText) != maxToolJSONBytes {
		t.Fatalf("truncatedJsonText len = %d, want %d", len(resp.TruncatedJsonText), maxToolJSONBytes)
	}
}

func TestJsonToolResultSuccess(t *testing.T) {
	t.Parallel()

	result, _, err := jsonToolResult(func() (string, error) {
		return marshalResponse(responseMeta{
			ToolName:      "demo",
			SchemaVersion: schemaVersion,
		})
	})
	if err != nil {
		t.Fatalf("jsonToolResult() error: %v", err)
	}
	if len(result.Content) != 1 {
		t.Fatalf("content len = %d, want 1", len(result.Content))
	}
}

func TestJsonToolResultPropagatesError(t *testing.T) {
	t.Parallel()

	_, _, err := jsonToolResult(func() (string, error) {
		return "", errors.New("boom")
	})
	if err == nil {
		t.Fatal("expected error propagation")
	}
}

