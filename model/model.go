package model

import (
	"context"
	"simple-agent/types"
)

type Model interface {
	NewStreaming(ctx context.Context, messages []types.Message) Stream
	NewSubagentStream(ctx context.Context, messages []types.Message) Stream
}

type Stream interface {
	Next() bool
	Err() error
	Current() string
	Response() string
	ToolCalls() []types.ToolCall
	StopReason() StopReason
}

type StopReason interface {
	String() string
	IsToolUse() bool
}
