package writefile

import (
	"context"
	"fmt"
	"os"

	"simple-agent/tools/base"
)

func New(cfg base.ToolConfig) base.Tool {
	return &writeFileTool{base.NewBaseTool(cfg.Name, cfg.Description, cfg.InputSchema)}
}

type writeFileTool struct{ base.BaseTool }

func (t *writeFileTool) Call(ctx context.Context, args map[string]any) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return "", fmt.Errorf("write_file: path must be a non-empty string")
	}
	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("write_file: content must be a string")
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return "ok", nil
}
