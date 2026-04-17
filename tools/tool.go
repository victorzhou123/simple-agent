package tools

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"simple-agent/tools/base"
	"simple-agent/tools/bash"
	editfile "simple-agent/tools/edit_file"
	readfile "simple-agent/tools/read_file"
	"simple-agent/tools/todo"
	writefile "simple-agent/tools/write_file"

	"github.com/anthropics/anthropic-sdk-go"
)

const (
	TOOL_NAME_BASH       = "bash"
	TOOL_NAME_READ_FILE  = "read_file"
	TOOL_NAME_WRITE_FILE = "write_file"
	TOOL_NAME_EDIT_FILE  = "edit_file"
	TOOL_NAME_TODO       = "todo"
)

// Re-export base types so callers only need to import "simple-agent/tools".
type (
	Tool       = base.Tool
	ToolConfig = base.ToolConfig
)

var NewBaseTool = base.NewBaseTool

//go:embed tools.json
var toolsJSON []byte

var factories = map[string]func(ToolConfig) Tool{
	TOOL_NAME_BASH:       bash.New,
	TOOL_NAME_READ_FILE:  readfile.New,
	TOOL_NAME_WRITE_FILE: writefile.New,
	TOOL_NAME_EDIT_FILE:  editfile.New,
	TOOL_NAME_TODO:       todo.New,
}

// Register allows external packages to add tools at startup.
func Register(name string, factory func(ToolConfig) Tool) {
	factories[name] = factory
}

// All holds every Tool in the order defined by tools.json.
var All []Tool

// Params is the Anthropic-SDK representation of All.
var Params []anthropic.ToolUnionParam

var index = map[string]Tool{}

func init() {
	var configs []ToolConfig
	if err := json.Unmarshal(toolsJSON, &configs); err != nil {
		panic("tools: failed to parse tools.json: " + err.Error())
	}
	for _, cfg := range configs {
		factory, ok := factories[cfg.Name]
		if !ok {
			panic(fmt.Sprintf("tools: no implementation for %q", cfg.Name))
		}
		t := factory(cfg)
		All = append(All, t)
		index[t.Name()] = t
		Params = append(Params, toAnthropicParam(t))
	}
}

// Call dispatches to the named tool.
func Call(name string, args map[string]any) (string, error) {
	t, ok := index[name]
	if !ok {
		return "", fmt.Errorf("tools: unknown tool %q", name)
	}
	return t.Call(args)
}

func toAnthropicParam(t Tool) anthropic.ToolUnionParam {
	return anthropic.ToolUnionParam{
		OfTool: &anthropic.ToolParam{
			Name:        t.Name(),
			Description: anthropic.String(t.Description()),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: t.Schema().Properties,
				Required:   t.Schema().Required,
			},
		},
	}
}
