package base

import "context"

type Tool interface {
	Name() string
	Description() string
	Schema() InputSchema
	Call(ctx context.Context, args map[string]any) (string, error)
}

type InputSchema struct {
	Properties map[string]any `json:"properties"`
	Required   []string       `json:"required"`
}

type ToolConfig struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"input_schema"`
}

type BaseTool struct {
	name        string
	description string
	schema      InputSchema
}

func NewBaseTool(name, description string, schema InputSchema) BaseTool {
	return BaseTool{name: name, description: description, schema: schema}
}

func (b BaseTool) Name() string        { return b.name }
func (b BaseTool) Description() string { return b.description }
func (b BaseTool) Schema() InputSchema { return b.schema }
