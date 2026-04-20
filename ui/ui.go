package ui

// UI 是 tea 包对外暴露的接口，Agent 调用它来驱动显示。
type UI interface {
	Run() error
	AppendChunk(text string)                                  // 推送一个流式文本块
	Done(finalContent string)                                 // 流式结束，传入完整内容
	Fail(err error)                                           // 出错
	ShowToolCall(toolName string, input map[string]any)       // 显示工具调用
	ShowToolResult(toolName string, output string, err error) // 显示工具结果
}
