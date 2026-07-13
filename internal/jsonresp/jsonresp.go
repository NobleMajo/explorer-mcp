package jsonresp

import (
	"encoding/json"
	"fmt"
)

const SchemaVersion = 1

type Meta struct {
	ToolName      string `json:"toolName"`
	SchemaVersion int    `json:"schemaVersion"`
}

func Marshal(v any) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal response: %w", err)
	}
	return string(data), nil
}
