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
	ctx context.Context, history []types.Message, question string,
) Stream {
	history = append(history, types.Message{Role: "user", Content: question})

	s := c.client.Messages.NewStreaming(ctx,
		anthropic.MessageNewParams{
			Model:     anthropic.Model(c.modelName),
			MaxTokens: 4096,
			System: []anthropic.TextBlockParam{
				{Text: c.systemPrompt.GetPrompt()},
			},
			Messages: buildParams(history),
		},
	)

	var assembled strings.Builder

	return &stream{
		stream:    s,
		assembled: assembled,
	}
}

type stream struct {
	assembled strings.Builder
	stream    *ssestream.Stream[anthropic.MessageStreamEventUnion]
}

func (s *stream) Next() bool {
	return s.stream.Next()
}

func (s *stream) Err() error {
	return s.stream.Err()
}

func (s *stream) Current() string {
	event := s.stream.Current()
	if e, ok := event.AsAny().(anthropic.ContentBlockDeltaEvent); ok {
		if d, ok := e.Delta.AsAny().(anthropic.TextDelta); ok {
			s.assembled.WriteString(d.Text)
			return d.Text
		}
	}
	return ""
}

func (s *stream) GetResponse() string {
	return s.assembled.String()
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
