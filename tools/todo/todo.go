package todo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"simple-agent/tools/base"
)

var (
	mu    sync.Mutex
	items []Item
)

type Item struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Status string `json:"status"`
}

func New(cfg base.ToolConfig) base.Tool {
	return &todoTool{base.NewBaseTool(cfg.Name, cfg.Description, cfg.InputSchema)}
}

type todoTool struct{ base.BaseTool }

func (t *todoTool) Call(ctx context.Context, args map[string]any) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	raw, ok := args["items"]
	if !ok {
		return "", fmt.Errorf("todo: missing required field 'items'")
	}

	b, err := json.Marshal(raw)
	if err != nil {
		return "", fmt.Errorf("todo: failed to marshal items: %w", err)
	}

	var incoming []Item
	if err := json.Unmarshal(b, &incoming); err != nil {
		return "", fmt.Errorf("todo: invalid items format: %w", err)
	}

	mu.Lock()
	defer mu.Unlock()

	index := make(map[string]int, len(items))
	for i, it := range items {
		index[it.ID] = i
	}

	for _, inc := range incoming {
		if i, exists := index[inc.ID]; exists {
			items[i] = inc
		} else {
			items = append(items, inc)
			index[inc.ID] = len(items) - 1
		}
	}

	var sb strings.Builder
	for _, it := range items {
		var mark string
		switch it.Status {
		case "completed":
			mark = "[x]"
		case "in_progress":
			mark = "[-]"
		default:
			mark = "[ ]"
		}
		fmt.Fprintf(&sb, "%s %s (%s)\n", mark, it.Text, it.ID)
	}
	return sb.String(), nil
}
