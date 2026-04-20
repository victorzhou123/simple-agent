package tools

import (
	"context"
	"fmt"

	"simple-agent/tools/base"

	"github.com/anthropics/anthropic-sdk-go"
)

// Re-export base types so callers only need to import "simple-agent/tools".
type (
	Tool       = base.Tool
	ToolConfig = base.ToolConfig
)

var NewBaseTool = base.NewBaseTool

// factories stores tool implementations, populated by init() in each tool package.
var factories = map[string]func(ToolConfig) Tool{}

// Register allows tool packages to register themselves at startup.
// Each tool package should call this in its init() function.
func Register(name string, factory func(ToolConfig) Tool) {
	if _, exists := factories[name]; exists {
		panic(fmt.Sprintf("tools: duplicate registration for %q", name))
	}
	factories[name] = factory
}

// All holds every Tool in the order defined by tools.json.
var All []Tool

// Params is the Anthropic-SDK representation of All.
var Params []anthropic.ToolUnionParam

var index = map[string]Tool{}

func Init(cfg []ToolConfig) {
	for _, cfg := range cfg {
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
func Call(ctx context.Context, name string, args map[string]any) (string, error) {
	t, ok := index[name]
	if !ok {
		return "", fmt.Errorf("tools: unknown tool %q", name)
	}
	return t.Call(ctx, args)
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
