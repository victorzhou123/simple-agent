package subagent

import (
	"context"
	"fmt"

	"simple-agent/model"
	"simple-agent/tools/base"
	"simple-agent/types"
	"simple-agent/ui"
)

func New(cfg base.ToolConfig, client model.Model, ui ui.UI, config Config,
	toolCaller func(ctx context.Context, name string, input map[string]any) (string, error)) base.Tool {
	return &subAgentTool{
		BaseTool:   base.NewBaseTool(cfg.Name, cfg.Description, cfg.InputSchema),
		client:     client,
		ui:         ui,
		config:     config,
		toolCaller: toolCaller,
	}
}

type subAgentTool struct {
	base.BaseTool

	client     model.Model
	ui         ui.UI
	config     Config
	toolCaller func(ctx context.Context, name string, input map[string]any) (string, error)
}

func (t *subAgentTool) Call(ctx context.Context, args map[string]any) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	prompt, ok := args["prompt"].(string)
	if !ok || prompt == "" {
		return "", fmt.Errorf("subagent: prompt must be a non-empty string")
	}
	// Initialize message history with user prompt
	messages := []types.Message{
		{Role: types.ROLE_USER, Content: prompt},
	}

	// Run agent loop (max 30 iterations for safety)
	for i := 0; i < t.config.MaxRounds; i++ {
		stream := t.client.NewSubagentStream(ctx, messages)
		for stream.Next() {
			// Silently consume streaming output to avoid UI pollution
			t.ui.AppendChunk(stream.Current())
		}

		if err := stream.Err(); err != nil {
			return "", err
		}

		finalContent := stream.Response()

		// function call
		toolCalls := stream.ToolCalls()
		if finalContent != "" || len(toolCalls) > 0 {
			msgType := types.MessageType("")
			if len(toolCalls) > 0 {
				msgType = types.TYPE_TOOL_USE
			}
			messages = append(messages, types.Message{
				Type:      msgType,
				Role:      types.ROLE_ASSISTANT,
				Content:   finalContent,
				ToolCalls: toolCalls,
			})
		}

		if !stream.StopReason().IsToolUse() {
			return finalContent, nil
		}

		// Execute each tool and collect results.
		results := make([]types.ToolResult, 0, len(toolCalls))
		for _, tc := range toolCalls {
			// 显示格式化的工具调用标题
			t.ui.ShowToolCall(tc.Name, tc.Input)

			out, err := t.toolCaller(ctx, tc.Name, tc.Input)

			// 显示格式化的工具执行结果
			t.ui.ShowToolResult(tc.Name, out, err)

			result := types.ToolResult{ToolUseID: tc.ID, Content: out}
			if err != nil {
				result.Content = err.Error()
				result.IsError = true
			}
			results = append(results, result)
		}

		messages = append(messages, types.Message{
			Type:        types.TYPE_TOOL_RESULT,
			Role:        types.ROLE_USER,
			ToolResults: results,
		})
	}

	return "(subagent reached iteration limit)", nil
}
