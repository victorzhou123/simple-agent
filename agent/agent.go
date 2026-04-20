package agent

import (
	"context"
	"time"

	"simple-agent/model"
	"simple-agent/tools"
	"simple-agent/types"
	"simple-agent/ui"
)

type Agent struct {
	client  model.Model
	ui      ui.UI
	history []types.Message
	config  Config
}

func New(client model.Model, ui ui.UI, cfg Config) *Agent {
	return &Agent{
		client: client,
		ui:     ui,
		config: cfg,
	}
}

func (a *Agent) OnSubmit(question string) {
	go func() {
		var usedTodo bool
		var roundsSinceTodo int

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.config.StreamTimeOut)*time.Second)
		defer cancel()

		for {
			if question != "" {
				a.history = append(a.history, types.Message{Role: types.ROLE_USER, Content: question})
			}

			stream := a.client.NewStreaming(ctx, a.history)
			for stream.Next() {
				a.ui.AppendChunk(stream.Current())
			}
			if err := stream.Err(); err != nil {
				a.ui.Fail(err)
				return
			}

			finalContent := stream.Response()
			toolCalls := stream.ToolCalls()

			if finalContent != "" || len(toolCalls) > 0 {
				msgType := types.MessageType("")
				if len(toolCalls) > 0 {
					msgType = types.TYPE_TOOL_USE
				}
				a.history = append(a.history, types.Message{
					Type:      msgType,
					Role:      types.ROLE_ASSISTANT,
					Content:   finalContent,
					ToolCalls: toolCalls,
				})
			}

			if !stream.StopReason().IsToolUse() {
				a.ui.Done(finalContent)
				return
			}

			// Execute each tool and collect results.
			results := make([]types.ToolResult, 0, len(toolCalls))
			for _, tc := range toolCalls {
				// 显示格式化的工具调用标题
				a.ui.ShowToolCall(tc.Name, tc.Input)

				if tc.Name == tools.TOOL_NAME_TODO {
					usedTodo = true
				}

				out, err := tools.Call(context.Background(), tc.Name, tc.Input)

				// 显示格式化的工具执行结果
				a.ui.ShowToolResult(tc.Name, out, err)

				result := types.ToolResult{ToolUseID: tc.ID, Content: out}
				if err != nil {
					result.Content = err.Error()
					result.IsError = true
				}
				results = append(results, result)
			}

			// todo reminder
			if usedTodo {
				if roundsSinceTodo >= 3 {
					a.history = append(a.history, types.BuildReminder("Update your todos."))
					roundsSinceTodo = 0
				} else {
					roundsSinceTodo++
				}
			}

			a.history = append(a.history, types.Message{
				Type:        types.TYPE_TOOL_RESULT,
				Role:        types.ROLE_USER,
				ToolResults: results,
			})
			question = ""
		}
	}()
}
