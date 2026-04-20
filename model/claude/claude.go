package claude

import (
	"context"
	"encoding/json"
	"strings"

	"simple-agent/model"
	"simple-agent/prompt"
	"simple-agent/tools"
	"simple-agent/types"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
)

type claudeCli struct {
	client       *anthropic.Client
	modelName    string
	systemPrompt prompt.SystemPrompt
}

func New(cfg Config, systemPrompt prompt.SystemPrompt) model.Model {
	opts := []option.RequestOption{option.WithAPIKey(cfg.ApiKey)}
	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}
	client := anthropic.NewClient(opts...)

	return &claudeCli{
		client:       &client,
		modelName:    cfg.Model,
		systemPrompt: systemPrompt,
	}
}

func (c *claudeCli) NewStreaming(
	ctx context.Context, messages []types.Message,
) model.Stream {
	// TODO error handling
	if err := ctx.Err(); err != nil {
		return nil
	}

	s := c.client.Messages.NewStreaming(ctx,
		anthropic.MessageNewParams{
			Model:     anthropic.Model(c.modelName),
			MaxTokens: 4096,
			System: []anthropic.TextBlockParam{
				{Text: c.systemPrompt.GetPrompt()},
			},
			Messages: buildParams(messages),
			Tools:    tools.AnthropicParams,
		},
	)

	return &stream{stream: s}
}

func (c *claudeCli) NewSubagentStream(ctx context.Context, messages []types.Message) model.Stream {
	// For subagent stream, we want to stop immediately when tool calls are generated, so we set max_tokens to 0 to disable the generation of response content.
	if err := ctx.Err(); err != nil {
		return nil
	}

	s := c.client.Messages.NewStreaming(ctx,
		anthropic.MessageNewParams{
			Model:     anthropic.Model(c.modelName),
			MaxTokens: 0,
			System: []anthropic.TextBlockParam{
				{Text: c.systemPrompt.GetPrompt()},
			},
			Messages: buildParams(messages),
			Tools:    tools.AnthropicSubParams,
		},
	)

	return &stream{stream: s}
}

type stream struct {
	textBuilder strings.Builder
	stream      *ssestream.Stream[anthropic.MessageStreamEventUnion]
	stopReason  model.StopReason

	inToolBlock     bool
	currentToolID   string
	currentToolName string
	currentJSON     strings.Builder
	toolCalls       []types.ToolCall
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
		_ = e
	case anthropic.ContentBlockStartEvent:
		switch block := e.ContentBlock.AsAny().(type) {
		case anthropic.ToolUseBlock:
			s.inToolBlock = true
			s.currentToolID = block.ID
			s.currentToolName = block.Name
			s.currentJSON.Reset()
		default:
			_ = block
			s.inToolBlock = false
		}
	case anthropic.ContentBlockDeltaEvent:
		switch d := e.Delta.AsAny().(type) {
		case anthropic.TextDelta:
			s.textBuilder.WriteString(d.Text)
			return d.Text
		case anthropic.InputJSONDelta:
			s.currentJSON.WriteString(d.PartialJSON)
		}
	case anthropic.ContentBlockStopEvent:
		if s.inToolBlock {
			var input map[string]any
			_ = json.Unmarshal([]byte(s.currentJSON.String()), &input)
			s.toolCalls = append(s.toolCalls, types.ToolCall{
				ID:    s.currentToolID,
				Name:  s.currentToolName,
				Input: input,
			})
			s.inToolBlock = false
		}
	case anthropic.MessageDeltaEvent:
		s.stopReason = stopReason(e.Delta.StopReason)
	case anthropic.MessageStopEvent:
		_ = e
	}
	return ""
}

func (s stream) Response() string {
	return s.textBuilder.String()
}

func (s stream) ToolCalls() []types.ToolCall {
	return s.toolCalls
}

func (s stream) StopReason() model.StopReason {
	return s.stopReason
}

type stopReason string // "end_turn", "max_tokens", "stop_sequence", "tool_use", "pause_turn"

func (s stopReason) String() string {
	return string(s)
}

func (s stopReason) IsToolUse() bool {
	return s.String() == string(types.TYPE_TOOL_USE)
}

func buildParams(history []types.Message) []anthropic.MessageParam {
	params := make([]anthropic.MessageParam, 0, len(history))
	for _, msg := range history {
		switch msg.Type {
		case types.TYPE_TOOL_RESULT:
			blocks := make([]anthropic.ContentBlockParamUnion, 0, len(msg.ToolResults))
			for _, tr := range msg.ToolResults {
				blocks = append(blocks, anthropic.NewToolResultBlock(tr.ToolUseID, tr.Content, tr.IsError))
			}
			params = append(params, anthropic.NewUserMessage(blocks...))

		case types.TYPE_TOOL_USE:
			blocks := make([]anthropic.ContentBlockParamUnion, 0, 1+len(msg.ToolCalls))
			if msg.Content != "" {
				blocks = append(blocks, anthropic.NewTextBlock(msg.Content))
			}
			for _, tc := range msg.ToolCalls {
				blocks = append(blocks, anthropic.NewToolUseBlock(tc.ID, tc.Input, tc.Name))
			}
			params = append(params, anthropic.NewAssistantMessage(blocks...))

		case types.TYPE_TEXT:
			if msg.Role == types.ROLE_USER {
				params = append(params, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
			} else {
				params = append(params, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)))
			}

		default:
			if msg.Role == types.ROLE_USER {
				params = append(params, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
			} else {
				params = append(params, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)))
			}
		}
	}
	return params
}
