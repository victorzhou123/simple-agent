package agent

import (
	"context"
	"fmt"

	"simple-agent/model"
	"simple-agent/tea"
	"simple-agent/tools"
	"simple-agent/types"
)

type Agent struct {
	client  model.Model
	ui      tea.UI
	history []types.Message
}

func New(client model.Model, ui tea.UI) *Agent {
	return &Agent{
		client: client,
		ui:     ui,
	}
}

func (a *Agent) OnSubmit(question string) {
	go func() {
		var usedTodo bool
		var roundsSinceTodo int

		for {
			if question != "" {
				a.history = append(a.history, types.Message{Role: types.ROLE_USER, Content: question})
			}

			stream := a.client.NewStreaming(context.Background(), a.history)
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
				a.ui.AppendChunk(fmt.Sprintf("\n\n[tool: %s]\n", tc.Name))

				if tc.Name == tools.TOOL_NAME_TODO {
					usedTodo = true
				}

				out, err := tools.Call(tc.Name, tc.Input)
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
