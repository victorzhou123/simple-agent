package agent

import (
	"context"

	"simple-agent/model"
	"simple-agent/tea"
	"simple-agent/types"
)

// Agent 持有业务状态：API 客户端、对话历史、模型名。
// 它通过 tea.UI 接口驱动显示，不直接依赖 bubbletea。
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

// OnSubmit 是注册给 tea.SubmitFunc 的回调。
// 它应当非阻塞：在内部启动 goroutine 发起流式请求，通过 ui 推送结果。
func (a *Agent) OnSubmit(question string) {
	go func() {
		for {
			if question != "" {
				a.history = append(a.history, types.Message{Role: "user", Content: question})
			}
			stream := a.client.NewStreaming(context.Background(), a.history)

			for stream.Next() {
				block := stream.Current()
				a.ui.AppendChunk(block)
			}

			// 检查错误
			if err := stream.Err(); err != nil {
				a.ui.Fail(err)
				return
			}

			finalContent := stream.Response()
			if finalContent != "" {
				a.history = append(a.history, types.Message{Role: "assistant", Content: finalContent})
			}
			a.ui.Done(finalContent)

			// 检查退出原因
			if !stream.StopReason().IsToolUse() {
				return
			}

			question = ""
		}
	}()
}
