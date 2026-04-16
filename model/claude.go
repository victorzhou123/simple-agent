package model

import (
	"context"
	"simple-agent/prompt"
	"simple-agent/types"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
)

type claudeCli struct {
	client       *anthropic.Client
	modelName    string
	systemPrompt prompt.SystemPrompt
}

func New(client *anthropic.Client, modelName string, systemPrompt prompt.SystemPrompt) Model {
	return &claudeCli{
		client:       client,
		modelName:    modelName,
		systemPrompt: systemPrompt,
	}
}

func (c *claudeCli) NewStreaming(
	ctx context.Context, messages []types.Message,
) Stream {
	s := c.client.Messages.NewStreaming(ctx,
		anthropic.MessageNewParams{
			Model:     anthropic.Model(c.modelName),
			MaxTokens: 4096,
			System: []anthropic.TextBlockParam{
				{Text: c.systemPrompt.GetPrompt()},
			},
			Messages: buildParams(messages),
		},
	)

	return &stream{stream: s}
}

type stream struct {
	assembled  strings.Builder
	stream     *ssestream.Stream[anthropic.MessageStreamEventUnion]
	stopReason StopReason
}

func (s *stream) Next() bool {
	return s.stream.Next()
}

func (s *stream) Err() error {
	return s.stream.Err()
}

func (s *stream) Current() string {
	switch e := s.stream.Current().AsAny().(type) {
	case anthropic.MessageStartEvent:
		_ = e // 可从 e.Message 获取 model、usage 等初始信息
	case anthropic.ContentBlockStartEvent:
		_ = e // 可从 e.ContentBlock 获取 block 类型
	case anthropic.ContentBlockDeltaEvent:
		switch d := e.Delta.AsAny().(type) {
		case anthropic.TextDelta:
			s.assembled.WriteString(d.Text)
			return d.Text
		case anthropic.InputJSONDelta:
			s.assembled.WriteString(d.PartialJSON)
			return d.PartialJSON
		}
	case anthropic.ContentBlockStopEvent:
		_ = e
	case anthropic.MessageDeltaEvent:
		s.stopReason = stopReason(e.Delta.StopReason)
	case anthropic.MessageStopEvent:
		_ = e
	}
	return ""
}

func (s stream) Response() string {
	return s.assembled.String()
}

func (s stream) StopReason() StopReason {
	return s.stopReason
}

type stopReason string // "end_turn", "max_tokens", "stop_sequence", "tool_use", "pause_turn"

func (s stopReason) String() string {
	return string(s)
}

func (s stopReason) IsToolUse() bool {
	return s.String() == "tool_use"
}

func buildParams(history []types.Message) []anthropic.MessageParam {
	params := make([]anthropic.MessageParam, 0, len(history))
	for _, msg := range history {
		if msg.Role == "user" {
			params = append(params, anthropic.NewUserMessage(
				anthropic.NewTextBlock(msg.Content),
			))
		} else {
			params = append(params, anthropic.NewAssistantMessage(
				anthropic.NewTextBlock(msg.Content),
			))
		}
	}
	return params
}
