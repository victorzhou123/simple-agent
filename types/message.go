package types

import "simple-agent/tools"

const (
	// type
	TYPE_TEXT        MessageType = "text"
	TYPE_TOOL_USE    MessageType = "tool_use"
	TYPE_TOOL_RESULT MessageType = "tool_result"

	// role
	ROLE_USER      MessageRole = "user"
	ROLE_ASSISTANT MessageRole = "assistant"
)

type MessageType string
type MessageRole string

type Message struct {
	Type        MessageType
	Role        MessageRole
	Content     string
	ToolCalls   []ToolCall
	ToolResults []ToolResult
}

type ToolCall struct {
	ID    string
	Name  string
	Input map[string]any
}

func (t ToolCall) IsToolTodo() bool {
	return t.Name == tools.TOOL_NAME_TODO
}

type ToolResult struct {
	ToolUseID string
	Content   string
	IsError   bool
}
