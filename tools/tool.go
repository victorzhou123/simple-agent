package tools

import (
	"context"
	"fmt"

	"simple-agent/model"
	"simple-agent/skill"
	"simple-agent/tools/base"
	"simple-agent/tools/bash"
	editfile "simple-agent/tools/edit_file"
	loadskill "simple-agent/tools/load_skill"
	readfile "simple-agent/tools/read_file"
	"simple-agent/tools/subagent"
	"simple-agent/tools/todo"
	writefile "simple-agent/tools/write_file"
	"simple-agent/ui"

	"github.com/anthropics/anthropic-sdk-go"
)

const (
	TOOL_NAME_BASH       = "bash"
	TOOL_NAME_READ_FILE  = "read_file"
	TOOL_NAME_WRITE_FILE = "write_file"
	TOOL_NAME_EDIT_FILE  = "edit_file"
	TOOL_NAME_TODO       = "todo"
	TOOL_NAME_SUBAGENT   = "subagent"
	TOOL_NAME_LOAD_SKILL = "load_skill"
)

// Re-export base types so callers only need to import "simple-agent/tools".
type (
	Tool       = base.Tool
	ToolConfig = base.ToolConfig
)

// Params is the Anthropic-SDK representation of All.
var AnthropicParams []anthropic.ToolUnionParam

// AnthropicSubParams is the Anthropic-SDK representation of all tools except subagent, which is used for subagent calls. We want to avoid infinite recursion of subagent calling itself.
var AnthropicSubParams []anthropic.ToolUnionParam

var toolIndex = map[string]Tool{}

func Init(cfg Config, cli model.Model, ui ui.UI, skillManager skill.SkillManager) {
	// tools registration
	for _, cf := range cfg.Tools {
		switch cf.Name {
		case TOOL_NAME_BASH:
			t := bash.New(cf)
			toolIndex[cf.Name] = t
			continue
		case TOOL_NAME_READ_FILE:
			t := readfile.New(cf)
			toolIndex[cf.Name] = t
			continue
		case TOOL_NAME_WRITE_FILE:
			t := writefile.New(cf)
			toolIndex[cf.Name] = t
			continue
		case TOOL_NAME_EDIT_FILE:
			t := editfile.New(cf)
			toolIndex[cf.Name] = t
			continue
		case TOOL_NAME_TODO:
			t := todo.New(cf)
			toolIndex[cf.Name] = t
			continue
		case TOOL_NAME_SUBAGENT:
			t := subagent.New(cf, cli, ui, cfg.SubagentConfig, Call)
			toolIndex[cf.Name] = t
			continue
		case TOOL_NAME_LOAD_SKILL:
			t := loadskill.New(cf, skillManager)
			toolIndex[cf.Name] = t
			continue
		default:
			panic(fmt.Sprintf("tools: unknown tool name %q in config", cf.Name))
		}
	}

	initAnthropicParams()
}

// Call dispatches to the named tool.
func Call(ctx context.Context, name string, args map[string]any) (string, error) {
	t, ok := toolIndex[name]
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

func initAnthropicParams() {
	// build anthropic params for all tools in toolIndex
	for _, t := range toolIndex {
		if t.Name() == TOOL_NAME_SUBAGENT {
			AnthropicParams = append(AnthropicParams, toAnthropicParam(t))
			continue
		}
		AnthropicSubParams = append(AnthropicSubParams, toAnthropicParam(t))
	}
}
